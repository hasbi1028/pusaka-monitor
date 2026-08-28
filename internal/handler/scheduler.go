package handler

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// Jadwal auto-scrape (WITA = UTC+8).
// Hanya 2x sehari untuk menghindari rate-limit/ban dari backend Pusaka:
//   - Pagi 07:01 → catch pegawai yang sudah masuk, flag yang belum masuk
//   - Sore 17:00 → catch yang sudah pulang, finalkan status
// Reference: auto-report.js guard harian (tiap jadwal 1x/hari).
var autoScrapeSchedules = []struct {
	Jam   int
	Menit int
	Label string
}{
	{7, 1, "pagi"},
	{17, 0, "sore"},
}

// lastRunHarian — catat tanggal terakhir tiap label jalan (guard anti-dobel)
var lastRunHarian = map[string]string{}

// StartScheduler — jalankan goroutine background:
//  1. Recover job stale tiap 60 detik (cegah stuck saat restart)
//  2. Auto-scrape di jam jadwal (07:01 pagi, 17:00 sore) — bukan polling
//
// Reference: auto-report.js checkAndTrigger() + scheduler.recoverStaleJobs().
func StartScheduler(db *gorm.DB) {
	go func() {
		// Recover stale tiap 60 dtk
		recoverTicker := time.NewTicker(60 * time.Second)
		defer recoverTicker.Stop()

		// Cek jadwal tiap 1 menit (cukup presisi untuk jam:menit)
		scheduleTicker := time.NewTicker(1 * time.Minute)
		defer scheduleTicker.Stop()

		log.Printf("[SCHEDULER] auto-scrape dijadwalkan: pagi 07:01, sore 17:00 WITA | recover-stale tiap 60s")

		for {
			select {
			case <-recoverTicker.C:
				n := databaseRecover(db)
				if n > 0 {
					log.Printf("[SCHEDULER] %d job stale di-reset ke pending", n)
				}
				// Cleanup: hapus job final (done/failed/cancelled) lebih dari 7 hari
				// agar DB tidak bengkak (auto-scrape 2x/hari × 35 pegawai).
				cutoff := time.Now().Add(-7 * 24 * time.Hour)
				db.Where("status IN ('done','failed','cancelled') AND created_at < ?", cutoff).
					Delete(&models.Job{})
			case <-scheduleTicker.C:
				wita := time.Now().UTC().Add(8 * time.Hour)
				todayStr := wita.Format("2006-01-02")
				for _, s := range autoScrapeSchedules {
					if wita.Hour() == s.Jam && wita.Minute() == s.Menit {
						// Guard harian: tiap label cuma 1x/hari
						if lastRunHarian[s.Label] == todayStr {
							continue
						}
						lastRunHarian[s.Label] = todayStr
						log.Printf("[SCHEDULER] auto-scrape %s (%02d:%02d WITA)", s.Label, s.Jam, s.Menit)
						autoScrapeAll(db)
					}
				}
			}
		}
	}()
}

// databaseRecover — wrapper ke database.RecoverStaleJobs
func databaseRecover(db *gorm.DB) int64 {
	return database.RecoverStaleJobs(db)
}

// autoScrapeAll — buat job untuk semua pegawai yang belum lengkap hari ini.
// Sama seperti tombol "Scrape Semua (Belum Lengkap)" tapi otomatis.
func autoScrapeAll(db *gorm.DB) {
	today := todayWITA()
	var pegawai []models.Pegawai
	db.Where("aktif = 1 AND password_pusaka != ''").Find(&pegawai)

	created := 0
	for _, p := range pegawai {
		var existing models.Absensi
		if db.Where("n_ip = ? AND instansi_id = ? AND tanggal = ?", p.NIP, p.InstansiID, today).First(&existing).Error == nil {
			if isLengkap(existing.JamMasuk, existing.JamPulang) {
				continue // sudah lengkap → skip
			}
		}
		job := models.Job{
			ID:           uuid.New().String(),
			InstansiID:   p.InstansiID,
			EmployeeID:   p.NIP,
			EmployeeName: p.Nama,
			Status:       "pending",
			MaxAttempts:  3,
		}
		if db.Create(&job).Error == nil {
			created++
		}
	}
	if created > 0 {
		log.Printf("[SCHEDULER] auto-scrape: %d job dibuat (belum lengkap)", created)
	}
}
