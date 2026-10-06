package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration
	CORSOrigin  string
}

func LoadENV() (Config, error) {
	_ = godotenv.Load()

	c := Config{
		HTTPAddr:    addressFromPort(),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		CORSOrigin:  env("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	var err error
	if c.JWTExpiry, err = time.ParseDuration(env("JWT_EXPIRY", "24h")); err != nil || c.JWTExpiry <= 0 {
		return Config{}, fmt.Errorf("invalid JWT_EXPIRY")
	}
	return c, nil
}

func addressFromPort() string {
	port := env("PORT", env("HTTP_ADDR", "8080"))
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
