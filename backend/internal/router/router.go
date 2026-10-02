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
	CommentHandler *handler.CommentHandler
	AdminHandler   *handler.AdminHandler
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
	api.Get("/channels/:id", deps.ChannelHandler.GetByID)
	api.Get("/posts", deps.PostHandler.List)
	api.Get("/posts/:id", deps.PostHandler.GetByID)
	api.Get("/posts/:id/comments", deps.CommentHandler.List)

	// Authenticated
	protected := api.Group("", authGuard)

	protected.Get("/me", func(c fiber.Ctx) error {
		claims := c.Locals("user").(*middleware.UserClaims)
		return c.JSON(claims)
	})
	protected.Post("/channels", deps.ChannelHandler.Create)
	protected.Post("/posts", deps.PostHandler.Create)
	protected.Put("/posts/:id", deps.PostHandler.Update)
	protected.Delete("/posts/:id", deps.PostHandler.Delete)
	protected.Post("/posts/:id/comments", deps.CommentHandler.Create)
	protected.Delete("/comments/:commentID", deps.CommentHandler.Delete)

	// Admin only
	admin := api.Group("", authGuard, middleware.RequireRole("admin"))

	admin.Get("/admin/users", deps.AdminHandler.ListUsers)
	admin.Get("/admin/stats", deps.AdminHandler.Stats)
	admin.Put("/admin/users/:id/roles", deps.AdminHandler.SetRoles)
	admin.Delete("/admin/users/:id", deps.AdminHandler.DeleteUser)
	admin.Delete("/channels/:id", deps.ChannelHandler.Delete)
}
