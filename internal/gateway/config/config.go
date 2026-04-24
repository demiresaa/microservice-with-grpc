package config

import (
	"os"
	"strings"
)

type Route struct {
	Path          string
	Target        string
	StripPrefix   string
	AuthRequired  bool
	RateLimitRPS  int
}

type GatewayConfig struct {
	Port       string
	JWTSecret  string
	Routes     []Route
}

func Load() *GatewayConfig {
	jwtSecret := getEnv("JWT_SECRET", "my-super-secret-key-change-in-production")

	return &GatewayConfig{
		Port:      getEnv("GATEWAY_PORT", "8000"),
		JWTSecret: jwtSecret,
		Routes: []Route{
			{
				Path:         "/api/auth",
				Target:       getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
				StripPrefix:  "/api",
				AuthRequired: false,
				RateLimitRPS: 10,
			},
			{
				Path:         "/api/orders",
				Target:       getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
				StripPrefix:  "/api",
				AuthRequired: true,
				RateLimitRPS: 20,
			},
			{
				Path:         "/api/health",
				Target:       getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
				StripPrefix:  "/api",
				AuthRequired: false,
				RateLimitRPS: 30,
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ParseRateLimitEnv(envKey string, defaultRPS int) int {
	v := os.Getenv(envKey)
	if v == "" {
		return defaultRPS
	}
	parts := strings.Split(v, ",")
	total := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			total++
		}
	}
	if total > 0 {
		return defaultRPS
	}
	return defaultRPS
}
