package services

import (
	"errors"
	"io"
	"path"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UploadService stores one or more uploaded parts into the caller's namespace.
// It owns the upload policy: per-request size cap, per-account storage quota,
// destination resolution and the "empty-name parts are skipped" rule, so the
// primary adapters stay thin.
type UploadService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
	index    ports.FileIndexRepository
	users    ports.UserRepository
	maxBytes int64 // <= 0 means unlimited
	bus      *events.Bus
}

func NewUploadService(fileRepo ports.FileRepository, scoper ports.PathScoper, index ports.FileIndexRepository, users ports.UserRepository, maxBytes int64) *UploadService {
	return &UploadService{fileRepo: fileRepo, scoper: scoper, index: index, users: users, maxBytes: maxBytes}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *UploadService) SetEventBus(bus *events.Bus) { s.bus = bus }

func (s *UploadService) Execute(user *models.User, parts []models.UploadPart) ([]models.FileUpload, error) {
	var uploads []models.FileUpload
	var execErrors []error
	remaining := s.maxBytes
	if remaining <= 0 {
		remaining = -1
	}
	budget := &byteBudget{remaining: remaining}
	qb, err := s.newQuotaBudget(user)
	if err != nil {
		return nil, err
	}

	for _, part := range parts {
		if part.Content == nil {
			continue
		}
		if part.Name == "" {
			part.Content.Close()
			continue
		}

		// A non-file multipart part arrives as an empty Name; everything else
		// is validated as a path, which also rejects traversal in either the
		// name or the destination.
		virtual := part.Name
		if part.Destination != "" {
			virtual = path.Join(part.Destination, part.Name)
		}
		fp, err := valueobjects.NewFilePath(virtual)
		if err != nil {
			part.Content.Close()
			execErrors = append(execErrors, err)
			continue
		}

		physical, err := s.scoper.WritePath(user, fp.Relative())
		if err != nil {
			part.Content.Close()
			execErrors = append(execErrors, err)
			continue
		}

		// A fully exhausted account is rejected before touching the filesystem,
		// so an existing file is never truncated by an over-quota write.
		if qb != nil && qb.remaining <= 0 {
			part.Content.Close()
			execErrors = append(execErrors, &domainerrors.QuotaExceededError{Action: "upload", Path: part.Name})
			continue
		}

		if err := s.ensureParent(physical); err != nil {
			part.Content.Close()
			execErrors = append(execErrors, err)
			continue
		}

		written, writeErr := s.fileRepo.WriteFile(physical, limited(part.Content, budget, qb))
		part.Content.Close()
		switch {
		case qb != nil && qb.exceeded:
			execErrors = append(execErrors, &domainerrors.QuotaExceededError{Action: "upload", Path: part.Name})
			continue
		case budget.exceeded:
			execErrors = append(execErrors, domainerrors.NewValidationError("files", nil, "upload exceeds size limit"))
			continue
		case writeErr != nil:
			execErrors = append(execErrors, writeErr)
			continue
		}

		virtualName := s.scoper.PhysicalToVirtual(user, physical)
		uploads = append(uploads, models.FileUpload{
			Filename: virtualName,
			Size:     written,
		})
		publishEventBytes(s.bus, events.TypeUpload, virtualName, user, written)
	}

	if len(execErrors) > 0 && len(uploads) == 0 {
		return nil, execErrors[0]
	}
	if len(uploads) == 0 {
		return nil, domainerrors.NewValidationError("files", nil, "no files uploaded")
	}
	return uploads, nil
}

// ensureParent creates the physical parent directory when it does not exist,
// so multi-level destinations land correctly without an explicit mkdir first.
func (s *UploadService) ensureParent(physical string) error {
	dir := path.Dir(path.Clean(physical))
	if dir == "." || dir == "/" {
		return nil
	}
	return s.fileRepo.CreateDirectory(dir)
}

// newQuotaBudget returns the per-account storage budget for this request: the
// account's quota minus its current usage. It returns nil when the account is
// unlimited or when there is no account (system view).
func (s *UploadService) newQuotaBudget(user *models.User) (*byteBudget, error) {
	if user == nil {
		return nil, nil
	}
	quota, err := s.users.GetQuotaBytes(user.Username)
	if err != nil {
		// A vanished account record has no quota to enforce.
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if quota <= 0 {
		return nil, nil
	}
	_, used, err := s.index.PrefixStats(s.scoper.PrivatePrefix(user.Username))
	if err != nil {
		return nil, err
	}
	remaining := quota - used
	if remaining < 0 {
		remaining = 0
	}
	return &byteBudget{remaining: remaining}, nil
}

// limited wraps content with every active budget (request size cap first,
// account quota second). A nil budget is ignored.
func limited(rc io.ReadCloser, budgets ...*byteBudget) io.ReadCloser {
	for _, b := range budgets {
		if b != nil {
			rc = b.limited(rc)
		}
	}
	return rc
}

// byteBudget tracks the remaining upload allowance across every part of a
// single request, so the total is capped rather than each individual file.
type byteBudget struct {
	remaining int64 // < 0 means unlimited
	exceeded  bool
}

// limited wraps content in a reader that flags exceeded once a part turns out
// larger than the budget that remains for the request.
func (b *byteBudget) limited(rc io.ReadCloser) io.ReadCloser {
	return &budgetReader{b: b, rc: rc, remaining: b.remaining}
}

type budgetReader struct {
	b         *byteBudget
	rc        io.ReadCloser
	remaining int64
	exceeded  bool
}

func (r *budgetReader) Read(p []byte) (int, error) {
	if r.exceeded || r.b.exceeded {
		return 0, io.EOF
	}
	if r.remaining < 0 {
		return r.rc.Read(p)
	}
	// Allow one byte past the budget so overflow is detectable without
	// misflagging a file that ends exactly on the limit.
	if int64(len(p)) > r.remaining+1 {
		p = p[:r.remaining+1]
	}
	n, err := r.rc.Read(p)
	if int64(n) > r.remaining {
		// Read into the overflow byte: the part exceeds the remaining budget.
		// Hand back only the allowed bytes (the extra one is discarded) and
		// stop, which aborts this part with a clean error upstream.
		r.exceeded = true
		r.b.exceeded = true
		r.remaining = 0
		return int(0), io.EOF
	}
	r.remaining -= int64(n)
	return n, err
}

func (r *budgetReader) Close() error { return r.rc.Close() }
