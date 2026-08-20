package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      App
	HTTP     HTTP
	Database Database
	Redis    Redis
	Storage  Storage
	Upload   Upload
	Auth     Auth
	CORS     CORS
}

type App struct {
	Env  string
	Name string
}

type HTTP struct {
	Host string
	Port string
}

type Database struct {
	URL string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

type Storage struct {
	Provider string // "local" | "s3"
	Local    LocalStorage
	S3       S3Storage
}

type LocalStorage struct {
	BaseDir string // on-disk root for the local storage driver
	BaseURL string // public base URL prefix for stored files
}

type S3Storage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	UseSSL          bool
	PublicBaseURL   string
}

type Upload struct {
	MaxSizeBytes int64
	AllowedTypes []string
}

type Auth struct {
	JWTSecret                string
	AccessTokenTTL           time.Duration
	RefreshTokenTTL          time.Duration
	RefreshTokenMaxAge       time.Duration
	ArgonMemoryKiB           uint32
	ArgonIterations          uint32
	ArgonParallelism         uint8
	ArgonSaltLength          int
	ArgonKeyLength           uint32
	SecureCookies            bool
	CrossOrigin              bool
	RefreshCookieDomain      string
	CookieName               string
}

type CORS struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: App{
			Env:  env("APP_ENV", "development"),
			Name: env("APP_NAME", "atchannel-api"),
		},
		HTTP: HTTP{
			Host: env("HTTP_HOST", "0.0.0.0"),
			Port: env("HTTP_PORT", "8080"),
		},
		Database: Database{
			URL: env("DATABASE_URL", "postgres://atchannel:atchannel@localhost:5432/atchannel?sslmode=disable"),
		},
		Redis: Redis{
			Addr:     env("REDIS_ADDR", "localhost:6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		Storage: Storage{
			Provider: env("STORAGE_PROVIDER", "local"),
			Local: LocalStorage{
				BaseDir: env("STORAGE_LOCAL_DIR", "/var/lib/atchannel/storage"),
				BaseURL: env("STORAGE_LOCAL_URL", "/files"),
			},
			S3: S3Storage{
				Endpoint:        env("S3_ENDPOINT", ""),
				AccessKeyID:     env("S3_ACCESS_KEY_ID", ""),
				SecretAccessKey: env("S3_SECRET_ACCESS_KEY", ""),
				Bucket:          env("S3_BUCKET", ""),
				Region:          env("S3_REGION", "us-east-1"),
				UseSSL:          envBool("S3_USE_SSL", true),
				PublicBaseURL:   env("S3_PUBLIC_BASE_URL", ""),
			},
		},
		Upload: Upload{
			MaxSizeBytes: envInt64("UPLOAD_MAX_BYTES", 10<<20),
			AllowedTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
		},
		Auth: Auth{
			JWTSecret:          env("JWT_SECRET", "dev-secret-change-me"),
			AccessTokenTTL:     envDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:    envDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
			RefreshTokenMaxAge: envDuration("REFRESH_TOKEN_ABSOLUTE_MAX_AGE", 90*24*time.Hour),
			ArgonMemoryKiB:     envUint32("ARGON_MEMORY_KIB", 65536),
			ArgonIterations:    envUint32("ARGON_ITERATIONS", 3),
			ArgonParallelism:   envUint8("ARGON_PARALLELISM", 1),
			ArgonSaltLength:    envInt("ARGON_SALT_LENGTH", 16),
			ArgonKeyLength:     envUint32("ARGON_KEY_LENGTH", 32),
			SecureCookies:      envBool("COOKIE_SECURE", false),
			CrossOrigin:        envBool("CROSS_ORIGIN", false),
			RefreshCookieDomain: env("REFRESH_COOKIE_DOMAIN", ""),
			CookieName:         env("REFRESH_COOKIE_NAME", "atch_refresh"),
		},
		CORS: CORS{
			AllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		},
	}

	if cfg.App.Env == "production" && cfg.Auth.JWTSecret == "dev-secret-change-me" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}
	if cfg.App.Env == "production" && (cfg.Storage.Provider == "s3" && cfg.Storage.S3.Bucket == "") {
		return nil, fmt.Errorf("S3_BUCKET must be set when STORAGE_PROVIDER=s3")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envInt64(key string, fallback int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func envUint32(key string, fallback uint32) uint32 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return fallback
	}
	return uint32(n)
}

func envUint8(key string, fallback uint8) uint8 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return fallback
	}
	return uint8(n)
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
