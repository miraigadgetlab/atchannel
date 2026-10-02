package router

import (
	"atchannel-backend/internal/config"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
)

type RouterDeps struct {
	Config         *config.Config
	AuthHandler    *handler.AuthHandler
	ChannelHandler *handler.ChannelHandler
	PostHandler    *handler.PostHandler
	CommentHandler *handler.CommentHandler
	AdminHandler   *handler.AdminHandler
	ProfileHandler *handler.ProfileHandler
}

func SetupRoutes(app *fiber.App, deps RouterDeps) {
	// CORS first so preflight OPTIONS is answered before any auth guard.
	app.Use(cors.New(cors.Config{
		AllowOrigins: deps.Config.AllowedOrigins,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodHead,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
		},
		AllowHeaders: []string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderAuthorization,
		},
		MaxAge: 3600,
	}))

	app.Use(middleware.RequestLogger())

	api := app.Group("/api/v1")

	authGuard := middleware.NewAuthMiddleware(middleware.AuthConfig{
		SecretKey:     deps.Config.JWTSecret,
		SigningMethod: jwt.SigningMethodHS256,
	})

	// Brute force protection on the public auth endpoints. Register counts
	// every attempt (mass signups are the attack), login only counts
	// failures (a legit user must never lock themselves out by succeeding).
	registerLimit := middleware.NewAuthRateLimiter(deps.Config.RateLimitMax, deps.Config.RateLimitWindow, false)
	loginLimit := middleware.NewAuthRateLimiter(deps.Config.RateLimitMax, deps.Config.RateLimitWindow, true)

	// Public
	api.Post("/auth/register", registerLimit, deps.AuthHandler.Register)
	api.Post("/auth/login", loginLimit, deps.AuthHandler.Login)
	api.Post("/auth/refresh", deps.AuthHandler.Refresh)
	api.Post("/auth/logout", deps.AuthHandler.Logout)

	api.Get("/channels", deps.ChannelHandler.List)
	api.Get("/channels/:id", deps.ChannelHandler.GetByID)
	api.Get("/posts", deps.PostHandler.List)
	api.Get("/posts/:id", deps.PostHandler.GetByID)
	api.Get("/posts/:id/comments", deps.CommentHandler.List)

	// Authenticated
	protected := api.Group("", authGuard)

	protected.Get("/me", deps.ProfileHandler.Me)
	protected.Put("/me", deps.ProfileHandler.Update)
	protected.Put("/me/password", deps.ProfileHandler.ChangePassword)
	protected.Post("/auth/logout-all", deps.AuthHandler.LogoutAll)
	protected.Post("/channels", deps.ChannelHandler.Create)
	protected.Put("/channels/:id", deps.ChannelHandler.Update)
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
