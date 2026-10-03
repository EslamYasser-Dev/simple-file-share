package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// socialError maps service errors for the social endpoints. Account-lookup
// misses become 404 here: unlike the public auth surface, social callers are
// already authenticated, so existence is no secret that follower lists and
// the feed do not already reveal.
func socialError(w http.ResponseWriter, err error) {
	if errors.Is(err, domainerrors.ErrUserNotFound) {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respondWithError(w, err)
}

// FollowsHandler manages open-follow edges on /api/follows: follow (POST),
// unfollow (DELETE), and reads (GET). All methods require authentication.
type FollowsHandler struct {
	svc   *services.FollowService
	audit *services.AuditService
}

func NewFollowsHandler(svc *services.FollowService) *FollowsHandler {
	return &FollowsHandler{svc: svc}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *FollowsHandler) SetAudit(a *services.AuditService) { h.audit = a }

func (h *FollowsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleFollow(w, r)
	case http.MethodDelete:
		h.handleUnfollow(w, r)
	case http.MethodGet:
		h.handleRead(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *FollowsHandler) handleFollow(w http.ResponseWriter, r *http.Request) {
	var req dto.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	me := actorName(r)
	if err := h.svc.Follow(me, req.Username); err != nil {
		socialError(w, err)
		return
	}
	h.audit.Record(services.AuditFollow, me, clientIP(r), req.Username, "")
	respondJSON(w, http.StatusCreated, dto.FollowResponse{Username: req.Username, Following: true})
}

func (h *FollowsHandler) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	var req dto.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	me := actorName(r)
	if err := h.svc.Unfollow(me, req.Username); err != nil {
		socialError(w, err)
		return
	}
	h.audit.Record(services.AuditUnfollow, me, clientIP(r), req.Username, "")
	respondJSON(w, http.StatusOK, dto.FollowResponse{Username: req.Username, Following: false})
}

func (h *FollowsHandler) handleRead(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// ?check=<username> answers "am I following them?".
	if target := q.Get("check"); target != "" {
		ok, err := h.svc.IsFollowing(actorName(r), target)
		if err != nil {
			socialError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, dto.FollowResponse{Username: target, Following: ok})
		return
	}
	subject := q.Get("user")
	if subject == "" {
		subject = actorName(r)
	}
	var usernames []string
	var err error
	if q.Get("direction") == "followers" {
		usernames, err = h.svc.Followers(subject)
	} else {
		usernames, err = h.svc.Following(subject)
	}
	if err != nil {
		socialError(w, err)
		return
	}
	if usernames == nil {
		usernames = []string{}
	}
	respondJSON(w, http.StatusOK, dto.UsernamesResponse{Usernames: usernames})
}

// VisibilityHandler manages per-file privacy on /api/visibility: set (PUT)
// and read (GET). Setting requires ownership; the level label itself reveals
// nothing about content. All methods require authentication.
type VisibilityHandler struct {
	svc   *services.VisibilityService
	audit *services.AuditService
}

func NewVisibilityHandler(svc *services.VisibilityService) *VisibilityHandler {
	return &VisibilityHandler{svc: svc}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *VisibilityHandler) SetAudit(a *services.AuditService) { h.audit = a }

func (h *VisibilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.handleSet(w, r)
	case http.MethodGet:
		h.handleGet(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *VisibilityHandler) handleSet(w http.ResponseWriter, r *http.Request) {
	var req dto.SetVisibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	vis, err := h.svc.Set(currentUser(r), req.Path, models.VisibilityLevel(req.Level), req.AllowStream)
	if err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditVisibilitySet, actorName(r), clientIP(r), req.Path, req.Level)
	respondJSON(w, http.StatusOK, dto.FromVisibility(vis))
}

func (h *VisibilityHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	owner := q.Get("owner")
	if owner == "" {
		owner = actorName(r)
	}
	vis, err := h.svc.Get(owner, q.Get("path"))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromVisibility(vis))
}

// SharedViewHandler streams another owner's file on
// GET /api/shared?owner=&path= after the visibility service re-checks the
// streaming gate. Safe types render inline (same allowlist as /api/view);
// anything else is forced to download. Requires authentication.
type SharedViewHandler struct {
	svc *services.VisibilityService
}

func NewSharedViewHandler(svc *services.VisibilityService) *SharedViewHandler {
	return &SharedViewHandler{svc: svc}
}

func (h *SharedViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	download, err := h.svc.Serve(currentUser(r), q.Get("owner"), q.Get("path"))
	if err != nil {
		socialError(w, err)
		return
	}
	if !download.Inline {
		serveDownload(w, r, download)
		return
	}
	serveInline(w, r, download)
}

// FeedHandler serves the upload timeline on /api/feed (GET only, cursor
// paged). The service re-applies audience rules on every read. All methods
// require authentication.
type FeedHandler struct {
	svc *services.TimelineService
}

func NewFeedHandler(svc *services.TimelineService) *FeedHandler {
	return &FeedHandler{svc: svc}
}

func (h *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	limit := 20
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := h.svc.Feed(currentUser(r), q.Get("cursor"), limit)
	if err != nil {
		respondWithError(w, err)
		return
	}
	resp := dto.FeedResponse{Events: dto.FromTimelineEvents(events)}
	if len(events) > 0 {
		resp.NextCursor = events[len(events)-1].ID
	}
	if resp.Events == nil {
		resp.Events = []dto.FeedItem{}
	}
	respondJSON(w, http.StatusOK, resp)
}
