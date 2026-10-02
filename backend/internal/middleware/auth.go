package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthConfig struct {
	SecretKey     []byte
	PublicKey     any
	SigningMethod jwt.SigningMethod
	ContextKey    string
}

func NewAuthMiddleware(cfg AuthConfig) fiber.Handler {
	if cfg.ContextKey == "" {
		cfg.ContextKey = "user"
	}
	if cfg.SigningMethod == nil {
		cfg.SigningMethod = jwt.SigningMethodHS256
	}

	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authorization header is missing",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or malformed authorization header format",
			})
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != cfg.SigningMethod.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			if cfg.PublicKey != nil {
				return cfg.PublicKey, nil
			}
			return cfg.SecretKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid, expired, or tampered token",
			})
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Failed to parse token claims",
			})
		}

		if claims.TokenType != "access" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "Invalid token type for this route",
			})
		}

		c.Locals(cfg.ContextKey, claims)

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
