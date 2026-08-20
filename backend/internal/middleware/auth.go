package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kosero/atchannel/backend/internal/models"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "userID"
	ctxUserRole ctxKey = "userRole"
	ctxBan      ctxKey = "ban"
)

func ContextUserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

func ContextUserRole(ctx context.Context) models.Role {
	v, _ := ctx.Value(ctxUserRole).(models.Role)
	return v
}

func ContextBan(ctx context.Context) bool {
	v, _ := ctx.Value(ctxBan).(bool)
	return v
}

// TokenValidator is implemented by the auth service; defining it here keeps
// the middleware free of a services import and breaks the import cycle.
type TokenValidator interface {
	ValidateAccessToken(token string) (userID string, role models.Role, err error)
}

type Auth struct {
	svc TokenValidator
}

func NewAuth(svc TokenValidator) *Auth {
	return &Auth{svc: svc}
}

// RequireAuth validates the Bearer token and injects the user identity.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		userID, role, err := a.svc.ValidateAccessToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxUserRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth parses the token when present but never rejects the request.
func (a *Auth) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token != "" {
			if userID, role, err := a.svc.ValidateAccessToken(token); err == nil {
				ctx := context.WithValue(r.Context(), ctxUserID, userID)
				ctx = context.WithValue(ctx, ctxUserRole, role)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

// RequireRole rejects requests whose authenticated role is below minimum.
func RequireRole(min models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := ContextUserRole(r.Context())
			if !roleAtLeast(role, min) {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func roleAtLeast(role, min models.Role) bool {
	rank := func(r models.Role) int {
		switch r {
		case models.RoleAdmin:
			return 3
		case models.RoleMod:
			return 2
		default:
			return 1
		}
	}
	return rank(role) >= rank(min)
}
