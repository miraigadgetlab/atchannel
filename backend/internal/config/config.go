package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	Port      string
	JWTSecret []byte

	// AllowedOrigins drives the CORS policy. Empty/unset means wildcard.
	AllowedOrigins []string

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
		Port:           ":" + port,
		JWTSecret:      []byte(secret),
		AllowedOrigins: parseOrigins(os.Getenv("CORS_ORIGINS")),
		AdminName:      os.Getenv("ADMIN_NAME"),
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
	}

	if cfg.AdminName == "" {
		cfg.AdminName = "admin"
	}

	if cfg.AdminEmail != "" && cfg.AdminPassword == "" {
		log.Print("WARNING: ADMIN_EMAIL is set but ADMIN_PASSWORD is empty, skipping admin bootstrap")
		cfg.AdminEmail = ""
	}

	log.Printf("CORS allowed origins: %v", cfg.AllowedOrigins)

	return cfg
}

// parseOrigins turns a comma separated CORS_ORIGINS value into a list,
// falling back to the wildcard when nothing usable was provided.
func parseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"*"}
	}

	var origins []string
	for _, part := range strings.Split(raw, ",") {
		origin := strings.TrimSuffix(strings.TrimSpace(part), "/")
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	if len(origins) == 0 {
		return []string{"*"}
	}

	return origins
}
