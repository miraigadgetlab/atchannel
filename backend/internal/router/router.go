package router

import (
	"atchannel-backend/internal/config"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type RouterDeps struct {
	Config         *config.Config
	AuthHandler    *handler.AuthHandler
	ChannelHandler *handler.ChannelHandler
	PostHandler    *handler.PostHandler
}

func SetupRoutes(app *fiber.App, deps RouterDeps) {
	app.Use(middleware.RequestLogger())

	api := app.Group("/api/v1")

	authGuard := middleware.NewAuthMiddleware(middleware.AuthConfig{
		SecretKey:     deps.Config.JWTSecret,
		SigningMethod: jwt.SigningMethodHS256,
	})

	// Public
	api.Post("/auth/register", deps.AuthHandler.Register)
	api.Post("/auth/login", deps.AuthHandler.Login)
	api.Post("/auth/refresh", deps.AuthHandler.Refresh)

	api.Get("/channels", deps.ChannelHandler.List)
	api.Get("/posts", deps.PostHandler.List)
	api.Get("/posts/:id", deps.PostHandler.GetByID)

	// Authenticated
	protected := api.Group("", authGuard)

	protected.Get("/me", func(c fiber.Ctx) error {
		claims := c.Locals("user").(*middleware.UserClaims)
		return c.JSON(claims)
	})
	protected.Post("/channels", deps.ChannelHandler.Create)
	protected.Post("/posts", deps.PostHandler.Create)
}
