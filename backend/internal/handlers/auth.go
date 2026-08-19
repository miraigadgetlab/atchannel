package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kosero/atchannel/backend/internal/config"
	"github.com/kosero/atchannel/backend/internal/services"
)

var (
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

type AuthHandler struct {
	svc *services.AuthService
	cfg *config.Auth
}

func NewAuthHandler(svc *services.AuthService, cfg *config.Auth) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !usernameRe.MatchString(req.Username) {
		writeError(w, http.StatusBadRequest, "username must be 3-32 chars (letters, digits, underscore)")
		return
	}
	if !emailRe.MatchString(req.Email) {
		writeError(w, http.StatusBadRequest, "invalid email address")
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 128 {
		writeError(w, http.StatusBadRequest, "password must be 8-128 chars")
		return
	}
	user, err := h.svc.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	tokens, err := h.svc.IssueTokens(r.Context(), user)
	if err != nil {
		writeErr(w, err)
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken)
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	tokens, err := h.svc.IssueTokens(r.Context(), user)
	if err != nil {
		writeErr(w, err)
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	raw := h.readRefreshToken(r)
	if raw == "" {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}
	tokens, user, err := h.svc.Refresh(r.Context(), raw)
	if err != nil {
		writeErr(w, err)
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	raw := h.readRefreshToken(r)
	if raw != "" {
		if err := h.svc.Logout(r.Context(), raw); err != nil {
			writeErr(w, err)
			return
		}
	}
	// Clear the cookie regardless of whether a token was present.
	sameSite := http.SameSiteStrictMode
	if h.cfg.CrossOrigin {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.SecureCookies,
		SameSite: sameSite,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// The auth middleware guarantees the identity; profile body is served
	// by the users handler for richer data. Here we just ack the session.
	writeJSON(w, http.StatusOK, map[string]string{"userID": userID(r), "role": role(r)})
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteStrictMode
	if h.cfg.CrossOrigin {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cfg.RefreshCookieDomain,
		HttpOnly: true,
		Secure:   h.cfg.SecureCookies,
		SameSite: sameSite,
		MaxAge:   int(h.cfg.RefreshTokenTTL.Seconds()),
		Expires:  time.Now().Add(h.cfg.RefreshTokenTTL),
	})
}

func (h *AuthHandler) readRefreshToken(r *http.Request) string {
	cookie, err := r.Cookie(h.cfg.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	return dec.Decode(v)
}
