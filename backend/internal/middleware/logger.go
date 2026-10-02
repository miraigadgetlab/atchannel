package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
)

func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		log.Printf("[%s] %s | Status: %d | Execution Time: %v",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration)

		return err
	}
}
