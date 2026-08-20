package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/services"
)

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, services.ErrUserExists):
		writeError(w, http.StatusConflict, "username or email already taken")
	case errors.Is(err, services.ErrRefreshTokenUsed):
		writeError(w, http.StatusUnauthorized, "refresh token reuse detected; please log in again")
	case errors.Is(err, services.ErrRefreshTokenExpired):
		writeError(w, http.StatusUnauthorized, "refresh token expired")
	case errors.Is(err, services.ErrRefreshTokenInvalid):
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
	case errors.Is(err, services.ErrBoardNotFound):
		writeError(w, http.StatusNotFound, "board not found")
	case errors.Is(err, services.ErrBoardExists):
		writeError(w, http.StatusConflict, "board already exists")
	case errors.Is(err, services.ErrInvalidBoardSlug):
		writeError(w, http.StatusBadRequest, "invalid board slug; use lowercase letters, numbers, hyphens, or underscores (max 20 chars)")
	case errors.Is(err, services.ErrBoardNameTooLong):
		writeError(w, http.StatusBadRequest, "board name too long")
	case errors.Is(err, services.ErrBoardDescTooLong):
		writeError(w, http.StatusBadRequest, "board description too long")
	case errors.Is(err, services.ErrThreadNotFound):
		writeError(w, http.StatusNotFound, "thread not found")
	case errors.Is(err, services.ErrReplyNotFound):
		writeError(w, http.StatusNotFound, "reply not found")
	case errors.Is(err, services.ErrThreadLocked):
		writeError(w, http.StatusForbidden, "thread is locked")
	case errors.Is(err, services.ErrBanned):
		writeError(w, http.StatusForbidden, "account is banned")
	case errors.Is(err, services.ErrTitleTooLong):
		writeError(w, http.StatusBadRequest, "title too long")
	case errors.Is(err, services.ErrBodyTooLong):
		writeError(w, http.StatusBadRequest, "body too long")
	case errors.Is(err, services.ErrNoImage):
		writeError(w, http.StatusBadRequest, "no image provided")
	case errors.Is(err, services.ErrImageTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "image exceeds maximum size")
	case errors.Is(err, services.ErrUnsupportedType):
		writeError(w, http.StatusUnsupportedMediaType, "unsupported image type")
	case errors.Is(err, services.ErrInvalidImage):
		writeError(w, http.StatusBadRequest, "invalid image data")
	case errors.Is(err, services.ErrInvalidReportTarget):
		writeError(w, http.StatusBadRequest, "invalid report")
	case errors.Is(err, services.ErrReportNotFound):
		writeError(w, http.StatusNotFound, "report not found")
	case errors.Is(err, services.ErrBanNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, repositories.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		slog.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
