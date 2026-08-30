package config

import (
	"log"
	"os"
)

type Config struct {
	Port          string
	DBPath        string
	JWTSecret     string
	AdminUser     string
	AdminPass     string
	EncryptionKey string
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("[FATAL] JWT_SECRET harus diset. Contoh: JWT_SECRET=rahasia-banget-xyz123")
	}
	adminPass := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("[FATAL] SUPERADMIN_PASSWORD harus diset. Contoh: SUPERADMIN_PASSWORD=Sp3derman!")
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		// Fallback ke JWT_SECRET kalau belum diset terpisah
		encryptionKey = jwtSecret
	}
	if len(encryptionKey) < 32 {
		log.Fatal("[FATAL] ENCRYPTION_KEY minimal 32 karakter (atau gunakan JWT_SECRET yang >=32)")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "data/presensi.db"),
		JWTSecret:     jwtSecret,
		AdminUser:     getEnv("SUPERADMIN_USERNAME", "admin"),
		AdminPass:     adminPass,
		EncryptionKey: encryptionKey,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
