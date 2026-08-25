package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipDirectory recursively zips a directory and writes to w.
// Designed to be used in a goroutine with io.Pipe().
func ZipDirectory(root string, w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer func() { _ = zipWriter.Close() }()

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		}
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			// Close eagerly per-file; a defer here would leak descriptors
			// until the entire walk completes.
			_, copyErr := io.Copy(writer, file)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}

		return nil
	})
}
