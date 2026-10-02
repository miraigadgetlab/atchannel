package config

import (
	"log"
	"os"
)

type Config struct {
	Port      string
	JWTSecret []byte
}

func LoadConfig() *Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	return &Config{
		Port:      ":" + port,
		JWTSecret: []byte(secret),
	}
}
