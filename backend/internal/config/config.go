package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port      string
	JWTSecret []byte

	// AllowedOrigins drives the CORS policy. Empty/unset means wildcard.
	AllowedOrigins []string

	// Brute-force protection for the public auth endpoints.
	// RateLimitMax/RateLimitWindow throttles login+register per client IP.
	// AccountLockMax/AccountLockWindow locks a single account after that
	// many consecutive failed logins, no matter where they come from.
	RateLimitMax      int
	RateLimitWindow   time.Duration
	AccountLockMax    int
	AccountLockWindow time.Duration

	// WriteLimitMax/WriteLimitWindow throttles content creation per user,
	// so one account cannot flood the site faster than a mod can clean up.
	WriteLimitMax    int
	WriteLimitWindow time.Duration

	// TrustedProxies is the allowlist of proxy addresses (IPs or CIDRs) whose
	// X-Forwarded-For header is believed. Empty means the app talks to clients
	// directly and any forwarded header is ignored — the safe default, since a
	// client can forge that header to escape per-IP rate limits.
	TrustedProxies []string

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
		Port:              ":" + port,
		JWTSecret:         []byte(secret),
		AllowedOrigins:    parseOrigins(os.Getenv("CORS_ORIGINS")),
		RateLimitMax:      envInt("RATE_LIMIT_MAX", 10),
		RateLimitWindow:   envDuration("RATE_LIMIT_WINDOW", time.Minute),
		AccountLockMax:    envInt("ACCOUNT_LOCK_MAX", 5),
		AccountLockWindow: envDuration("ACCOUNT_LOCK_WINDOW", 15*time.Minute),
		WriteLimitMax:     envInt("WRITE_LIMIT_MAX", 30),
		WriteLimitWindow:  envDuration("WRITE_LIMIT_WINDOW", time.Minute),
		TrustedProxies:    parseList(os.Getenv("TRUSTED_PROXIES")),
		AdminName:         os.Getenv("ADMIN_NAME"),
		AdminEmail:        os.Getenv("ADMIN_EMAIL"),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
	}

	if cfg.AdminName == "" {
		cfg.AdminName = "admin"
	}

	if cfg.AdminEmail != "" && cfg.AdminPassword == "" {
		log.Print("WARNING: ADMIN_EMAIL is set but ADMIN_PASSWORD is empty, skipping admin bootstrap")
		cfg.AdminEmail = ""
	}

	log.Printf("CORS allowed origins: %v", cfg.AllowedOrigins)
	log.Printf("rate limits: %d attempts/%s per IP, account lock %d failures/%s, writes %d/%s per user",
		cfg.RateLimitMax, cfg.RateLimitWindow, cfg.AccountLockMax, cfg.AccountLockWindow,
		cfg.WriteLimitMax, cfg.WriteLimitWindow)
	if len(cfg.TrustedProxies) == 0 {
		log.Print("trusted proxies: none (X-Forwarded-For ignored)")
	} else {
		log.Printf("trusted proxies: %v", cfg.TrustedProxies)
	}

	return cfg
}

// envInt reads an integer override, falling back on empty or invalid values.
func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		log.Printf("WARNING: invalid %s=%q, using default %d", name, raw, fallback)
		return fallback
	}

	return value
}

// envDuration reads a Go duration override ("30s", "15m"), falling back on
// empty or invalid values.
func envDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		log.Printf("WARNING: invalid %s=%q, using default %s", name, raw, fallback)
		return fallback
	}

	return value
}

// parseList turns a comma separated value into a trimmed list, dropping blanks.
func parseList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
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
