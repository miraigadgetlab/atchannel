package router

import (
	"atchannel-backend/internal/config"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type RouterDeps struct {
	Config      *config.Config
	AuthHandler *handler.AuthHandler
}

func SetupRoutes(app *fiber.App, deps RouterDeps) {
	app.Use(middleware.RequestLogger())

	api := app.Group("/api/v1")

	api.Post("/auth/register", deps.AuthHandler.Register)
	api.Post("/auth/login", deps.AuthHandler.Login)
	api.Post("/auth/refresh", deps.AuthHandler.Refresh)

	authGuard := middleware.NewAuthMiddleware(middleware.AuthConfig{
		SecretKey:     deps.Config.JWTSecret,
		SigningMethod: jwt.SigningMethodHS256,
	})

	protected := api.Group("", authGuard)
	protected.Get("/me", func(c fiber.Ctx) error {
		claims := c.Locals("user").(*middleware.UserClaims)
		return c.JSON(claims)
	})
}
