package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Address         string
	LegacyDSN       string
	FrontendDistDir string
}

func Load() Config {
	projectRoot := discoverProjectRoot()

	return Config{
		Address:         envOrDefault("RXSG_ADDR", ":8080"),
		LegacyDSN:       envOrDefault("RXSG_MYSQL_DSN", "root:@tcp(127.0.0.1:3306)/bloodwar?charset=utf8&parseTime=true&loc=Local"),
		FrontendDistDir: envOrDefault("RXSG_FRONTEND_DIST", filepath.Join(projectRoot, "frontend", "dist")),
	}
}

func discoverProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
