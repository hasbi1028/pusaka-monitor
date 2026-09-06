package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

func Init(dbURL string) *gorm.DB {
	var dialector gorm.Dialector

	// Support both PostgreSQL URL and individual env vars
	if dbURL != "" {
		dialector = postgres.Open(dbURL)
	} else {
		// Build DSN from individual env vars
		host := getEnv("POSTGRES_HOST", "localhost")
		port := getEnv("POSTGRES_PORT", "5432")
		user := getEnv("POSTGRES_USER", "postgres")
		pass := getEnv("POSTGRES_PASSWORD", "")
		name := getEnv("POSTGRES_DB", "pusaka_monitor")
		sslmode := getEnv("POSTGRES_SSLMODE", "disable")

		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Makassar",
			host, port, user, pass, name, sslmode,
		)
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		// Info "record not found" saat worker idle dibuang agar tak
		// membanjiri log; error beneran + slow query tetap tercatat.
		Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			logger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// Connection pooling untuk production
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying DB:", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	// Auto migrate
	err = db.AutoMigrate(
		&models.Instansi{},
		&models.User{},
		&models.Pegawai{},
		&models.Absensi{},
		&models.Job{},
		&models.Session{},
		&models.Setting{},
		&models.ApprovalLog{},
		&models.Schedule{},
		&models.RecapLog{},
		&models.Cuti{},
		&models.HariLibur{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create unique index if not exists (PostgreSQL syntax)
	db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_indexes WHERE indexname = 'uni_absensi'
			) THEN
				-- Clean duplicates first
				DELETE FROM absensis
				WHERE id IN (
					SELECT id FROM (
						SELECT id,
						   ROW_NUMBER() OVER (
							PARTITION BY n_ip, tanggal, instansi_id
							ORDER BY scraped_at DESC, id DESC
						   ) AS rn
						FROM absensis
					) WHERE rn > 1
				);
				CREATE UNIQUE INDEX uni_absensi ON absensis(n_ip, tanggal, instansi_id);
			END IF;
		END $$;
	`)
	log.Println("[DB] PostgreSQL connected + migrations applied")

	return db
}

// RecoverStaleJobs — reset job 'running' yang tertinggal (worker mati /
// pm2 restart) kembali ke 'pending' agar tidak stuck selamanya.
func RecoverStaleJobs(db *gorm.DB) int64 {
	res := db.Model(&models.Job{}).
		Where("status = ? AND claimed_at IS NOT NULL AND claimed_at < NOW() - INTERVAL '10 minutes'", "running").
		Updates(map[string]interface{}{
			"status":      "pending",
			"worker_id":   "",
			"claimed_at":  nil,
			"error":       "recovered (stale)",
		})
	return res.RowsAffected
}

func SeedSuperAdmin(db *gorm.DB, username, password string) {
	var count int64
	db.Model(&models.User{}).Where("role = ?", "superadmin").Count(&count)
	if count > 0 {
		return
	}

	hash, err := HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	user := models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		Role:         "superadmin",
		Aktif:        true,
	}
	db.Create(&user)
	log.Printf("[SEED] Superadmin created: %s", username)
}

// CleanupExpiredSessions — hapus session yang sudah expired
func CleanupExpiredSessions(db *gorm.DB) int64 {
	result := db.Where("expires_at < NOW()").Delete(&models.Session{})
	return result.RowsAffected
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
