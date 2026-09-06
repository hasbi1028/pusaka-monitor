package backup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// StartAutoBackup menjalankan backup otomatis sesuai interval
func StartAutoBackup(db *gorm.DB, interval time.Duration) {
	go func() {
		// Jalankan backup pertama kali setelah 1 jam
		time.Sleep(1 * time.Hour)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			backup(db)
		}
	}()
	log.Printf("[BACKUP] Auto-backup dijadwalkan tiap %v (PostgreSQL)", interval)
}

func backup(db *gorm.DB) {
	dir := "backups"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[BACKUP] Gagal buat direktori: %v", err)
		return
	}
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("pusaka_%s.sql", timestamp)
	dst := filepath.Join(dir, filename)

	// Get PostgreSQL connection info from env
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	dbname := getEnv("POSTGRES_DB", "pusaka_monitor")

	// Use pg_dump for backup
	cmd := exec.Command("pg_dump",
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbname,
		"-f", dst,
		"--no-owner",
		"--no-privileges",
	)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", getEnv("POSTGRES_PASSWORD", "")),
	)

	if err := cmd.Run(); err != nil {
		log.Printf("[BACKUP] Gagal backup: %v", err)
		return
	}

	// Get file size
	info, err := os.Stat(dst)
	if err != nil {
		log.Printf("[BACKUP] Gagal baca file: %v", err)
		return
	}

	log.Printf("[BACKUP] OK: %s (%d bytes)", dst, info.Size())

	// Cleanup: hapus backup lebih dari 7 hari
	entries, _ := os.ReadDir(dir)
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	removed := 0
	for _, e := range entries {
		info, err := e.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, e.Name()))
			removed++
		}
	}
	if removed > 0 {
		log.Printf("[BACKUP] Cleanup: %d backup lama dihapus", removed)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
