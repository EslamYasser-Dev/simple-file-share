package s3

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// FileRepository implements FileRepository, VersionRepository and StorageMover
// on an S3-compatible object store. Directories are trailing-slash markers;
// file history lives under `<file>.versions/<n>` keys.
type FileRepository struct {
	client      *Client
	versionKeep int
}

var (
	_ ports.FileRepository    = (*FileRepository)(nil)
	_ ports.VersionRepository = (*FileRepository)(nil)
	_ ports.StorageMover      = (*FileRepository)(nil)
)

func NewFileRepository(settings ports.S3Settings) *FileRepository {
	return &FileRepository{client: NewClient(settings)}
}

// SetVersionKeep bounds how many historical snapshots are retained per file.
// A value of 0 (the default) keeps every snapshot.
func (r *FileRepository) SetVersionKeep(n int) {
	if n < 0 {
		n = 0
	}
	r.versionKeep = n
}

func (r *FileRepository) resolve(path string) (string, error) {
	cleaned := strings.TrimPrefix(path, "/")
	if cleaned == "." {
		cleaned = ""
	}
	for _, segment := range strings.Split(cleaned, "/") {
		if segment == ".." {
			return "", errors.NewValidationError("path", path, "path traversal detected")
		}
	}
	return cleaned, nil
}

func shouldSkipListing(relKey string) bool {
	if relKey == "" {
		return true
	}
	base := path.Base(relKey)
	if strings.HasPrefix(relKey, ".file-share") || strings.Contains(relKey, "/.file-share/") {
		return true
	}
	if strings.HasSuffix(base, ".versions") || strings.Contains(relKey, ".versions/") {
		return true
	}
	return strings.HasPrefix(base, ".")
}

func joinPath(dir, name string) string {
	dir = strings.Trim(dir, "/")
	if dir == "" {
		return name
	}
	return dir + "/" + name
}

func (r *FileRepository) versionPrefix(fileKey string) string {
	return fileKey + ".versions/"
}

func (r *FileRepository) versionNumbers(fileKey string) ([]int, error) {
	var nums []int
	prefix := r.versionPrefix(fileKey)
	err := r.client.listAll(listOptions{prefix: prefix}, func(batch *listBucketResult) error {
		for _, obj := range batch.Contents {
			if strings.HasSuffix(obj.Key, "/") {
				continue
			}
			if n, err := strconv.Atoi(path.Base(obj.Key)); err == nil && n > 0 {
				nums = append(nums, n)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Ints(nums)
	return nums, nil
}

func (r *FileRepository) versionCount(fileKey string) (uint64, error) {
	nums, err := r.versionNumbers(fileKey)
	if err != nil {
		return 0, err
	}
	return uint64(len(nums)), nil
}

func (r *FileRepository) ListDirectory(dirPath string) ([]*models.FileInfo, error) {
	cleaned, err := r.resolve(dirPath)
	if err != nil {
		return nil, err
	}
	prefix := r.client.objectKey(cleaned)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	seen := map[string]struct{}{}
	var files []*models.FileInfo
	err = r.client.listAll(listOptions{prefix: prefix, delimiter: "/"}, func(batch *listBucketResult) error {
		for _, p := range batch.CommonPrefixes {
			rel := r.client.fromObjectKey(p.Prefix)
			if shouldSkipListing(rel) {
				continue
			}
			name := strings.TrimSuffix(path.Base(rel), "/")
			if name == "" || name == "." {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			files = append(files, &models.FileInfo{
				Name:  name,
				Path:  joinPath(cleaned, name),
				IsDir: true,
			})
		}
		for _, obj := range batch.Contents {
			rel := r.client.fromObjectKey(obj.Key)
			if strings.HasSuffix(obj.Key, "/") {
				name := strings.TrimSuffix(path.Base(rel), "/")
				if name == "" || name == "." {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				files = append(files, &models.FileInfo{
					Name:  name,
					Path:  joinPath(cleaned, name),
					IsDir: true,
				})
				continue
			}
			if shouldSkipListing(rel) {
				continue
			}
			name := path.Base(rel)
			if _, ok := seen[name]; ok {
				continue
			}
			version, _ := r.versionCount(r.client.objectKey(rel))
			files = append(files, &models.FileInfo{
				Name:     name,
				Path:     joinPath(cleaned, name),
				Size:     obj.Size,
				IsDir:    false,
				Modified: obj.LastModified,
				Version:  version,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})
	return files, nil
}

func (r *FileRepository) GetFileInfo(filePath string) (*models.FileInfo, error) {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return nil, err
	}
	if cleaned == "" {
		return &models.FileInfo{Name: "", Path: "", IsDir: true}, nil
	}
	key := r.client.objectKey(cleaned)

	if isDir, err := r.isDir(cleaned, key); err != nil {
		return nil, err
	} else if isDir {
		return &models.FileInfo{Name: path.Base(cleaned), Path: cleaned, IsDir: true}, nil
	}

	size, mod, err := r.client.headObject(key)
	if err != nil {
		if isNotFound(err) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	version, _ := r.versionCount(key)
	return &models.FileInfo{
		Name:     path.Base(cleaned),
		Path:     cleaned,
		Size:     size,
		IsDir:    false,
		Modified: mod,
		Version:  version,
	}, nil
}

func (r *FileRepository) isDir(cleaned, key string) (bool, error) {
	if cleaned == "" {
		return true, nil
	}
	marker := key
	if !strings.HasSuffix(marker, "/") {
		marker += "/"
	}
	if _, _, err := r.client.headObject(marker); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	found := false
	err := r.client.listAll(listOptions{prefix: marker, delimiter: "/"}, func(batch *listBucketResult) error {
		if len(batch.Contents) > 0 || len(batch.CommonPrefixes) > 0 {
			found = true
			return errStopList
		}
		return nil
	})
	if err != nil && err != errStopList {
		return false, err
	}
	return found, nil
}

var errStopList = fmt.Errorf("stop list")

func (r *FileRepository) IsDirectory(filePath string) (bool, error) {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return false, err
	}
	return r.isDir(cleaned, r.client.objectKey(cleaned))
}

func (r *FileRepository) FileExists(filePath string) (bool, error) {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return false, err
	}
	if cleaned == "" {
		return true, nil
	}
	key := r.client.objectKey(cleaned)
	if _, _, err := r.client.headObject(key); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	return r.isDir(cleaned, key)
}

func (r *FileRepository) ServeFile(filePath string) (io.ReadCloser, string, error) {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return nil, "", err
	}
	key := r.client.objectKey(cleaned)
	rc, _, _, err := r.client.getObject(key)
	if err != nil {
		if isNotFound(err) {
			return nil, "", errors.ErrNotFound
		}
		return nil, "", err
	}
	return rc, path.Base(cleaned), nil
}

func (r *FileRepository) CreateDirectory(dirPath string) error {
	cleaned, err := r.resolve(dirPath)
	if err != nil {
		return err
	}
	if cleaned == "" {
		return nil
	}
	// Create markers for every intermediate segment (mirrors MkdirAll).
	parts := strings.Split(cleaned, "/")
	acc := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if acc == "" {
			acc = part
		} else {
			acc = acc + "/" + part
		}
		key := r.client.objectKey(acc) + "/"
		if err := r.client.putObjectBytes(key, nil); err != nil {
			return err
		}
	}
	return nil
}

func (r *FileRepository) DeletePath(filePath string) error {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return err
	}
	if cleaned == "" {
		return errors.NewValidationError("path", filePath, "cannot delete storage root")
	}
	key := r.client.objectKey(cleaned)

	isDirectory, err := r.isDir(cleaned, key)
	if err != nil {
		return err
	}
	if isDirectory {
		return r.deleteUnder(key)
	}
	_ = r.deleteUnder(r.versionPrefix(key))
	return r.client.deleteObject(key)
}

func (r *FileRepository) deleteUnder(prefix string) error {
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	var keys []string
	err := r.client.listAll(listOptions{prefix: prefix}, func(batch *listBucketResult) error {
		for _, obj := range batch.Contents {
			keys = append(keys, obj.Key)
		}
		return nil
	})
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		if err := r.client.deleteObject(k); err != nil && !isNotFound(err) {
			return err
		}
	}
	// Remove the directory marker itself when prefix is a dir key without slash.
	if !strings.HasSuffix(strings.TrimSuffix(prefix, "/"), "/") {
		// prefix already ends with / after normalize; also try bare key.
	}
	bare := strings.TrimSuffix(prefix, "/")
	if bare != "" {
		if err := r.client.deleteObject(bare); err != nil && !isNotFound(err) {
			return err
		}
	}
	return r.client.deleteObject(prefix)
}

func (r *FileRepository) WriteFile(filePath string, reader io.ReadCloser) (int64, error) {
	defer func() { _ = reader.Close() }()

	cleaned, err := r.resolve(filePath)
	if err != nil {
		return 0, err
	}
	key := r.client.objectKey(cleaned)

	if _, _, err := r.client.headObject(key); err == nil {
		if err := r.snapshot(key); err != nil {
			return 0, err
		}
	} else if !isNotFound(err) {
		return 0, err
	}

	if parent := path.Dir(cleaned); parent != "." && parent != "/" && parent != "" {
		_ = r.client.putObjectBytes(r.client.objectKey(parent)+"/", nil)
	}

	tmp, err := os.CreateTemp("", "s3-upload-*")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	h := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(tmp, h), reader)
	if copyErr != nil {
		tmp.Close()
		return written, copyErr
	}
	payloadHash := hex.EncodeToString(h.Sum(nil))
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		return written, err
	}
	putErr := r.client.putObjectStream(key, tmp, written, payloadHash)
	closeErr := tmp.Close()
	if putErr != nil {
		return written, putErr
	}
	if closeErr != nil {
		return written, closeErr
	}
	r.pruneVersions(key)
	return written, nil
}

func (r *FileRepository) snapshot(key string) error {
	nums, err := r.versionNumbers(key)
	if err != nil {
		return err
	}
	next := 1
	if len(nums) > 0 {
		next = nums[len(nums)-1] + 1
	}
	return r.client.copyObject(key, r.versionPrefix(key)+strconv.Itoa(next))
}

func (r *FileRepository) pruneVersions(key string) {
	if r.versionKeep <= 0 {
		return
	}
	nums, err := r.versionNumbers(key)
	if err != nil || len(nums) <= r.versionKeep {
		return
	}
	for _, n := range nums[:len(nums)-r.versionKeep] {
		_ = r.client.deleteObject(r.versionPrefix(key) + strconv.Itoa(n))
	}
}

func (r *FileRepository) ListVersions(filePath string) ([]*models.FileInfo, error) {
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return nil, err
	}
	exists, err := r.FileExists(cleaned)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.ErrNotFound
	}
	if isDir, err := r.IsDirectory(cleaned); err != nil {
		return nil, err
	} else if isDir {
		return nil, nil
	}

	key := r.client.objectKey(cleaned)
	nums, err := r.versionNumbers(key)
	if err != nil {
		return nil, err
	}
	versions := make([]*models.FileInfo, 0, len(nums))
	for _, n := range nums {
		size, mod, err := r.client.headObject(r.versionPrefix(key) + strconv.Itoa(n))
		if err != nil {
			continue
		}
		versions = append(versions, &models.FileInfo{
			Name:     path.Base(cleaned),
			Path:     cleaned,
			Size:     size,
			IsDir:    false,
			Modified: mod,
			Version:  uint64(n),
		})
	}
	return versions, nil
}

func (r *FileRepository) ServeVersion(filePath string, n int) (io.ReadCloser, string, error) {
	if n < 1 {
		return nil, "", errors.ErrNotFound
	}
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return nil, "", err
	}
	key := r.client.objectKey(cleaned)
	rc, _, _, err := r.client.getObject(r.versionPrefix(key) + strconv.Itoa(n))
	if err != nil {
		if isNotFound(err) {
			return nil, "", errors.ErrNotFound
		}
		return nil, "", err
	}
	return rc, path.Base(cleaned), nil
}

func (r *FileRepository) RestoreVersion(filePath string, n int) error {
	if n < 1 {
		return errors.ErrNotFound
	}
	cleaned, err := r.resolve(filePath)
	if err != nil {
		return err
	}
	key := r.client.objectKey(cleaned)
	vkey := r.versionPrefix(key) + strconv.Itoa(n)

	rc, _, _, err := r.client.getObject(vkey)
	if err != nil {
		if isNotFound(err) {
			return errors.ErrNotFound
		}
		return err
	}
	defer rc.Close()

	if _, _, err := r.client.headObject(key); err == nil {
		if err := r.snapshot(key); err != nil {
			return err
		}
	} else if !isNotFound(err) {
		return err
	}

	tmp, err := os.CreateTemp("", "s3-restore-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	h := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmp, h), rc)
	if err != nil {
		tmp.Close()
		return err
	}
	payloadHash := hex.EncodeToString(h.Sum(nil))
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		return err
	}
	putErr := r.client.putObjectStream(key, tmp, size, payloadHash)
	closeErr := tmp.Close()
	if putErr != nil {
		return putErr
	}
	r.pruneVersions(key)
	return closeErr
}

// MovePath renames a physical path (username rename: users/a → users/b).
func (r *FileRepository) MovePath(oldPath, newPath string) error {
	src, err := r.resolve(oldPath)
	if err != nil {
		return err
	}
	dst, err := r.resolve(newPath)
	if err != nil {
		return err
	}
	srcKey := r.client.objectKey(src)
	dstKey := r.client.objectKey(dst)

	if _, _, err := r.client.headObject(srcKey); err == nil {
		return r.client.copyObject(srcKey, dstKey)
	} else if !isNotFound(err) {
		return err
	}

	srcPrefix := srcKey
	if !strings.HasSuffix(srcPrefix, "/") {
		srcPrefix += "/"
	}
	var keys []string
	if err := r.client.listAll(listOptions{prefix: srcPrefix}, func(batch *listBucketResult) error {
		for _, obj := range batch.Contents {
			keys = append(keys, obj.Key)
		}
		return nil
	}); err != nil {
		return err
	}
	if len(keys) == 0 {
		return r.CreateDirectory(dst)
	}
	dstPrefix := dstKey
	if !strings.HasSuffix(dstPrefix, "/") {
		dstPrefix += "/"
	}
	for _, k := range keys {
		rel := strings.TrimPrefix(k, srcPrefix)
		if err := r.client.copyObject(k, dstPrefix+rel); err != nil {
			return err
		}
	}
	return r.deleteUnder(srcPrefix)
}

func (r *FileRepository) ZipDirectory(root string) (io.ReadCloser, error) {
	cleaned, err := r.resolve(root)
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	go func() {
		err := r.writeZip(cleaned, pw)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()
	return pr, nil
}

func (r *FileRepository) writeZip(dir string, w io.Writer) error {
	prefix := r.client.objectKey(dir)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	type entry struct {
		key  string
		size int64
		mod  int64
		dir  bool
	}
	var entries []entry
	err := r.client.listAll(listOptions{prefix: prefix}, func(batch *listBucketResult) error {
		for _, p := range batch.CommonPrefixes {
			entries = append(entries, entry{key: p.Prefix, dir: true})
		}
		for _, obj := range batch.Contents {
			rel := r.client.fromObjectKey(obj.Key)
			if shouldSkipListing(rel) && !strings.HasSuffix(obj.Key, "/") {
				continue
			}
			if strings.HasSuffix(obj.Key, "/") {
				entries = append(entries, entry{key: obj.Key, dir: true})
				continue
			}
			entries = append(entries, entry{key: obj.Key, size: obj.Size, mod: obj.LastModified.Unix()})
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })

	zw := zip.NewWriter(w)
	for _, e := range entries {
		rel := strings.TrimPrefix(e.key, prefix)
		if rel == "" {
			continue
		}
		hdr := &zip.FileHeader{Name: rel}
		if e.dir {
			if !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
			hdr.Method = zip.Store
			if _, err := zw.CreateHeader(hdr); err != nil {
				return err
			}
			continue
		}
		rc, _, mod, err := r.client.getObject(e.key)
		if err != nil {
			return err
		}
		hdr.Method = zip.Deflate
		if !mod.IsZero() {
			hdr.SetModTime(mod)
		}
		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(fw, rc)
		closeErr := rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return zw.Close()
}

// Walk lists every file and directory for index rebuild. rootDir is ignored.
func (r *FileRepository) Walk(rootDir string) ([]*models.FileInfo, error) {
	_ = rootDir
	dirSeen := map[string]struct{}{}
	var entries []*models.FileInfo

	addDirChain := func(rel string) {
		parts := strings.Split(strings.Trim(rel, "/"), "/")
		acc := ""
		for i, part := range parts {
			if part == "" {
				continue
			}
			if acc == "" {
				acc = part
			} else {
				acc = acc + "/" + part
			}
			last := i == len(parts)-1
			if _, ok := dirSeen[acc]; ok {
				continue
			}
			// Intermediate segments are directories; a trailing marker is too.
			if !last || strings.HasSuffix(rel, "/") {
				dirSeen[acc] = struct{}{}
				entries = append(entries, &models.FileInfo{
					Name:  part,
					Path:  acc,
					IsDir: true,
				})
			}
		}
	}

	err := r.client.listAll(listOptions{}, func(batch *listBucketResult) error {
		for _, p := range batch.CommonPrefixes {
			rel := r.client.fromObjectKey(p.Prefix)
			if shouldSkipListing(rel) {
				continue
			}
			addDirChain(rel)
		}
		for _, obj := range batch.Contents {
			rel := r.client.fromObjectKey(obj.Key)
			if shouldSkipListing(rel) {
				continue
			}
			if strings.HasSuffix(obj.Key, "/") {
				addDirChain(rel)
				continue
			}
			// Ensure parent dirs exist in the listing.
			if parent := path.Dir(rel); parent != "." && parent != "/" {
				addDirChain(parent + "/")
			}
			version, _ := r.versionCount(obj.Key)
			entries = append(entries, &models.FileInfo{
				Name:     path.Base(rel),
				Path:     rel,
				Size:     obj.Size,
				IsDir:    false,
				Modified: obj.LastModified,
				Version:  version,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// WalkTextDocuments streams indexable text objects for full-text rebuild.
func (r *FileRepository) WalkTextDocuments(rootDir string) ([]ports.TextDocument, error) {
	entries, err := r.Walk(rootDir)
	if err != nil {
		return nil, err
	}
	var docs []ports.TextDocument
	for _, e := range entries {
		if e.IsDir || !policy.IsIndexableTextPath(e.Path) {
			continue
		}
		rc, _, err := r.ServeFile(e.Path)
		if err != nil {
			continue
		}
		content, ok := policy.ReadIndexableText(e.Path, rc)
		_ = rc.Close()
		if !ok {
			continue
		}
		docs = append(docs, ports.TextDocument{Path: e.Path, Content: content})
	}
	return docs, nil
}
