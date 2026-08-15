package handlers

import (
	"net/http"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/services"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := URLParam(r, "username")
	user, err := h.svc.GetByUsername(r.Context(), username)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type updateProfileRequest struct {
	AvatarURL string `json:"avatarUrl"`
	Bio       string `json:"bio"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user, err := h.svc.CurrentUser(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	var req updateProfileRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.svc.UpdateProfile(r.Context(), user.ID, req.AvatarURL, req.Bio)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// currentUserLoader resolves the full user for the authenticated request.
type currentUserLoader struct {
	users *services.UserService
}

func (l *currentUserLoader) load(r *http.Request) (*models.User, error) {
	return l.users.CurrentUser(r.Context())
}
