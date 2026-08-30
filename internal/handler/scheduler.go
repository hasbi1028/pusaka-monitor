package handler

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// lastRunHarian — catat tanggal terakhir tiap label jalan (guard anti-dobel)
var lastRunHarian = map[string]string{}

// StartScheduler — jalankan goroutine background:
//  1. Recover job stale tiap 60 detik
//  2. Auto-scrape sesuai jadwal dari database
func StartScheduler(db *gorm.DB) {
	go func() {
		// Recover stale tiap 60 dtk
		recoverTicker := time.NewTicker(60 * time.Second)
		defer recoverTicker.Stop()

		// Cek jadwal tiap 30 detik
		scheduleTicker := time.NewTicker(30 * time.Second)
		defer scheduleTicker.Stop()

		log.Printf("[SCHEDULER] auto-scrape dari database | recover-stale tiap 60s")

		for {
			select {
			case <-recoverTicker.C:
				n := databaseRecover(db)
				if n > 0 {
					log.Printf("[SCHEDULER] %d job stale di-reset ke pending", n)
				}
				// Cleanup: hapus job final lebih dari 7 hari
				cutoff := time.Now().Add(-7 * 24 * time.Hour)
				db.Where("status IN ('done','failed','cancelled') AND created_at < ?", cutoff).
					Delete(&models.Job{})

			case <-scheduleTicker.C:
				runSchedules(db)
			}
		}
	}()
}

// runSchedules — cek semua jadwal aktif dari DB, jalankan jika waktunya
func runSchedules(db *gorm.DB) {
	wita := time.Now().UTC().Add(8 * time.Hour)
	todayStr := wita.Format("2006-01-02")
	currentHour := wita.Hour()
	currentMin := wita.Minute()

	var schedules []models.Schedule
	db.Where("aktif = 1").Find(&schedules)

	for _, s := range schedules {
		if s.Jam == currentHour && s.Menit == currentMin {
			// Guard harian: tiap ID cuma 1x/hari
			runKey := s.ID
			if lastRunHarian[runKey] == todayStr {
				continue
			}
			lastRunHarian[runKey] = todayStr
			log.Printf("[SCHEDULER] auto-scrape %s (%02d:%02d WITA, mode=%s)", s.Label, s.Jam, s.Menit, s.Mode)
			autoScrapeByMode(db, s.Mode)
		}
	}
}

// databaseRecover — wrapper ke database.RecoverStaleJobs
func databaseRecover(db *gorm.DB) int64 {
	return database.RecoverStaleJobs(db)
}

// autoScrapeByMode — buat job sesuai mode jadwal
func autoScrapeByMode(db *gorm.DB, mode string) {
	today := todayWITA()
	var pegawai []models.Pegawai
	db.Where("aktif = 1 AND password_pusaka != ''").Find(&pegawai)

	created := 0
	for _, p := range pegawai {
		// Cek absensi hari ini
		var existing models.Absensi
		hasAbsen := db.Where("n_ip = ? AND instansi_id = ? AND tanggal = ?", p.NIP, p.InstansiID, today).First(&existing).Error == nil

		switch mode {
		case "belum_masuk":
			if hasAbsen && existing.JamMasuk != "" && existing.JamMasuk != "-" {
				continue
			}
		case "belum_pulang":
			if !(hasAbsen && existing.JamMasuk != "" && existing.JamMasuk != "-" && (existing.JamPulang == "" || existing.JamPulang == "-")) {
				continue
			}
		default: // all — scrape semua
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
		log.Printf("[SCHEDULER] auto-scrape: %d job dibuat (mode=%s)", created, mode)
	}
}
