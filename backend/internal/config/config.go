package config

import (
	"log"
	"os"
)

type Config struct {
	Port      string
	JWTSecret []byte

	// Bootstrap admin account. Seeding only happens when both
	// AdminEmail and AdminPassword are provided.
	AdminName     string
	AdminEmail    string
	AdminPassword string
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

	cfg := &Config{
		Port:          ":" + port,
		JWTSecret:     []byte(secret),
		AdminName:     os.Getenv("ADMIN_NAME"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}

	if cfg.AdminName == "" {
		cfg.AdminName = "admin"
	}

	if cfg.AdminEmail != "" && cfg.AdminPassword == "" {
		log.Print("WARNING: ADMIN_EMAIL is set but ADMIN_PASSWORD is empty, skipping admin bootstrap")
		cfg.AdminEmail = ""
	}

	return cfg
}
