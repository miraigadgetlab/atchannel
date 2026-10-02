package main

import (
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

	authHandler := handler.NewAuthHandler(tokenService, userService)
	channelHandler := handler.NewChannelHandler(channelService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	app := fiber.New()

	router.SetupRoutes(app, router.RouterDeps{
		Config:         cfg,
		AuthHandler:    authHandler,
		ChannelHandler: channelHandler,
		PostHandler:    postHandler,
		CommentHandler: commentHandler,
	})

	log.Printf("listening on %s", cfg.Port)
	log.Fatal(app.Listen(cfg.Port))
}
