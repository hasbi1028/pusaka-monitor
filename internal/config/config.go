package config

import (
	"log"
	"os"
)

type Config struct {
	Port          string
	DBURL         string // PostgreSQL connection URL (optional, overrides individual vars)
	DBPath        string // Legacy: SQLite path (ignored when DBURL is set)
	JWTSecret     string
	AdminUser     string
	AdminPass     string
	EncryptionKey string
	// PostgreSQL individual vars (used when DBURL is empty)
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	// GoWA WhatsApp gateway (rekap gambar harian)
	GowaBaseURL string
	GowaUser    string
	GowaPass    string
	GowaDevice  string
	RecapGroup  string
	// Telegram fallback (kirim bila WA gagal)
	TelegramBotToken string
	TelegramChatID   string
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
		DBURL:         os.Getenv("DATABASE_URL"),
		DBPath:        getEnv("DB_PATH", "data/presensi.db"), // Legacy, ignored when DATABASE_URL set
		JWTSecret:     jwtSecret,
		AdminUser:     getEnv("SUPERADMIN_USERNAME", "admin"),
		AdminPass:     adminPass,
		EncryptionKey: encryptionKey,
		// PostgreSQL
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       getEnv("POSTGRES_DB", "pusaka_monitor"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		// GoWA
		GowaBaseURL: getEnv("GOWA_BASE_URL", ""),
		GowaUser:    getEnv("GOWA_USER", "admin"),
		GowaPass:    os.Getenv("GOWA_PASS"),
		GowaDevice:  getEnv("GOWA_DEVICE", "mtsnbot"),
		RecapGroup:  getEnv("RECAP_GROUP", "120363409303983377@g.us"),
		// Telegram
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
