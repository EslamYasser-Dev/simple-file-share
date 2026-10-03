package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// UserSearchHandler serves GET /api/users/search?q= — username lookup for
// calling and sharing. Usernames only, requester excluded, capped. Requires
// authentication.
type UserSearchHandler struct {
	svc *services.DirectoryService
}

func NewUserSearchHandler(svc *services.DirectoryService) *UserSearchHandler {
	return &UserSearchHandler{svc: svc}
}

func (h *UserSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	usernames, err := h.svc.Search(currentUser(r), r.URL.Query().Get("q"))
	if err != nil {
		respondWithError(w, err)
		return
	}
	if usernames == nil {
		usernames = []string{}
	}
	respondJSON(w, http.StatusOK, dto.UsernamesResponse{Usernames: usernames})
}
