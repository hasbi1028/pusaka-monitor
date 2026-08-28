package config

import "os"

type Config struct {
	Port       string
	DBPath     string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "data/presensi.db"),
		JWTSecret: getEnv("JWT_SECRET", "pusaka-monitor-secret-2026"),
		AdminUser: getEnv("SUPERADMIN_USERNAME", "admin"),
		AdminPass: getEnv("SUPERADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
