package main

import (
	"context"
	"log"

	"atchannel-backend/internal/config"
	"atchannel-backend/internal/db"
	"atchannel-backend/internal/handler"
	"atchannel-backend/internal/router"
	"atchannel-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.LoadConfig()

	db.Connect()
	db.Migrate()

	tokenService := service.NewTokenService(cfg.JWTSecret, "atchannel")
	userService := service.NewUserService(db.DB)
	channelService := service.NewChannelService(db.DB)
	postService := service.NewPostService(db.DB)
	commentService := service.NewCommentService(db.DB)
	statsService := service.NewStatsService(db.DB)

	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		admin, err := userService.EnsureAdmin(context.Background(), cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword)
		if err != nil {
			log.Fatalf("admin bootstrap failed: %v", err)
		}
		log.Printf("admin account ready: %s <%s> roles=%v", admin.Name, admin.Email, admin.Roles)
	}

	authHandler := handler.NewAuthHandler(tokenService, userService)
	channelHandler := handler.NewChannelHandler(channelService, postService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	adminHandler := handler.NewAdminHandler(userService, statsService)
	profileHandler := handler.NewProfileHandler(userService)

	app := fiber.New()

	router.SetupRoutes(app, router.RouterDeps{
		Config:         cfg,
		AuthHandler:    authHandler,
		ChannelHandler: channelHandler,
		PostHandler:    postHandler,
		CommentHandler: commentHandler,
		AdminHandler:   adminHandler,
		ProfileHandler: profileHandler,
	})

	log.Printf("listening on %s", cfg.Port)
	log.Fatal(app.Listen(cfg.Port))
}
