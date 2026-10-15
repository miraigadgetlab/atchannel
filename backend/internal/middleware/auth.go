package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthConfig struct {
	SecretKey     []byte
	PublicKey     any
	SigningMethod jwt.SigningMethod
	Issuer        string
	ContextKey    string
}

func normalizeAuthConfig(cfg AuthConfig) AuthConfig {
	if cfg.ContextKey == "" {
		cfg.ContextKey = "user"
	}
	if cfg.SigningMethod == nil {
		cfg.SigningMethod = jwt.SigningMethodHS256
	}
	return cfg
}

// authError carries the HTTP status an authentication failure should surface.
type authError struct {
	code    int
	message string
}

func (e *authError) Error() string { return e.message }

// authenticate parses the bearer token and, on success, stores the claims in
// the context. It never writes a response; callers decide how to handle the
// error. A missing header is reported as a 401 authError, which lets optional
// callers distinguish "no credentials" from "bad credentials".
func authenticate(c fiber.Ctx, cfg AuthConfig) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return &authError{fiber.StatusUnauthorized, "Authorization header is missing"}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return &authError{fiber.StatusUnauthorized, "Invalid or malformed authorization header format"}
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{cfg.SigningMethod.Alg()}),
	}
	if cfg.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(cfg.Issuer))
	}

	token, err := jwt.ParseWithClaims(parts[1], &UserClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != cfg.SigningMethod.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		if cfg.PublicKey != nil {
			return cfg.PublicKey, nil
		}
		return cfg.SecretKey, nil
	}, parserOptions...)

	if err != nil || !token.Valid {
		return &authError{fiber.StatusUnauthorized, "Invalid, expired, or tampered token"}
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return &authError{fiber.StatusUnauthorized, "Failed to parse token claims"}
	}

	if claims.TokenType != "access" {
		return &authError{fiber.StatusForbidden, "Invalid token type for this route"}
	}

	c.Locals(cfg.ContextKey, claims)

	return nil
}

func NewAuthMiddleware(cfg AuthConfig) fiber.Handler {
	cfg = normalizeAuthConfig(cfg)

	return func(c fiber.Ctx) error {
		if err := authenticate(c, cfg); err != nil {
			var ae *authError
			if errors.As(err, &ae) {
				return c.Status(ae.code).JSON(fiber.Map{
					"error":   http.StatusText(ae.code),
					"message": ae.message,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authentication failed",
			})
		}

		return c.Next()
	}
}

// OptionalAuth attaches the caller's claims when a valid bearer token is
// present, and otherwise lets the request through as anonymous. A malformed or
// expired token is treated as anonymous rather than an error, so a stale token
// never breaks a public read.
func OptionalAuth(cfg AuthConfig) fiber.Handler {
	cfg = normalizeAuthConfig(cfg)

	return func(c fiber.Ctx) error {
		if c.Get(fiber.HeaderAuthorization) != "" {
			_ = authenticate(c, cfg)
		}
		return c.Next()
	}
}

func RequireRole(requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims, ok := c.Locals("user").(*UserClaims)
		if !ok || claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		userRoles := make(map[string]bool)
		for _, r := range claims.Roles {
			userRoles[r] = true
		}

		for _, role := range requiredRoles {
			if userRoles[role] {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "Forbidden",
			"message": "You do not have permission to access this resource",
		})
	}
}
