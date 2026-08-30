package database

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

func Init(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// WAL mode untuk performa + safety (cegah corrupt saat crash)
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")

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
		&models.Cuti{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Migration: pastikan unique index (n_ip, tanggal, instansi_id) ada
	// untuk mencegah duplikat absensi (bug: baris ABDILLAH muncul 3x).
	if !indexExists(db, "absensis", "uni_absensi") {
		// Hapus duplikat: keep 1 row per (n_ip, tanggal, instansi_id).
		// Pakai CTE + row_number (SQLite 3.25+).
		db.Exec(`
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
			)
		`)
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uni_absensi ON absensis(n_ip, tanggal, instansi_id)")
		log.Println("[MIGRATE] unique index uni_absensi dibuat + duplikat dibersihkan")
	}

	return db
}

// RecoverStaleJobs — reset job 'running' yang tertinggal (worker mati /
// pm2 restart) kembali ke 'pending' agar tidak stuck selamanya.
// Reference: db.recoverStaleJobs() dijalankan tiap 60 detik.
func RecoverStaleJobs(db *gorm.DB) int64 {
	res := db.Model(&models.Job{}).
		Where("status = ? AND claimed_at IS NOT NULL AND claimed_at < datetime('now','-10 minutes')", "running").
		Updates(map[string]interface{}{
			"status":      "pending",
			"worker_id":   "",
			"claimed_at":  nil,
			"error":       "recovered (stale)",
		})
	return res.RowsAffected
}

// indexExists — cek apakah index ada di tabel
func indexExists(db *gorm.DB, table, index string) bool {
	var cnt int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=? AND tbl_name=?", index, table).Scan(&cnt)
	return cnt > 0
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
		Username:     username,
		PasswordHash: hash,
		Role:         "superadmin",
		Aktif:        true,
	}
	db.Create(&user)
	log.Printf("[SEED] Superadmin created: %s", username)
}
