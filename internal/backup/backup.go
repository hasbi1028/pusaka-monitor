package backup

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// StartAutoBackup menjalankan backup otomatis sesuai interval
func StartAutoBackup(dbPath string, interval time.Duration) {
	go func() {
		// Jalankan backup pertama kali setelah 1 jam
		time.Sleep(1 * time.Hour)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			backup(dbPath)
		}
	}()
	log.Printf("[BACKUP] Auto-backup dijadwalkan tiap %v", interval)
}

func backup(dbPath string) {
	dir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[BACKUP] Gagal buat direktori: %v", err)
		return
	}
	timestamp := time.Now().Format("2006-01-02_150405")
	dst := filepath.Join(dir, fmt.Sprintf("presensi_%s.db", timestamp))

	src, err := os.Open(dbPath)
	if err != nil {
		log.Printf("[BACKUP] Gagal buka DB: %v", err)
		return
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		log.Printf("[BACKUP] Gagal buat file backup: %v", err)
		return
	}
	defer dstFile.Close()

	n, err := io.Copy(dstFile, src)
	if err != nil {
		log.Printf("[BACKUP] Gagal copy: %v", err)
		return
	}

	log.Printf("[BACKUP] OK: %s (%d bytes)", dst, n)

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
