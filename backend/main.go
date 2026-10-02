package main

import (
	"atchannel-backend/internal/db"
	"atchannel-backend/internal/middleware"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	db.Connect()
	db.Migrate()

	app := fiber.New()

	app.Use(middleware.RequestLogger())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hai")
	})

	log.Fatal(app.Listen(":3000"))
}
