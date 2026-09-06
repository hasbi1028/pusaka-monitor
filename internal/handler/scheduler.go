package handler

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
	"github.com/hasbiawal/pusaka-monitor/internal/recap"
)

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

// runSchedules — cek semua jadwal aktif dari DB, jalankan jika waktunya.
// Klaim atomik via kolom last_run: aman dari proses ganda & restart
// (hanya 1 proses yang dapat RowsAffected=1 per jadwal per hari).
func runSchedules(db *gorm.DB) {
	wita := time.Now().UTC().Add(8 * time.Hour)
	todayStr := wita.Format("2006-01-02")
	currentHour := wita.Hour()
	currentMin := wita.Minute()

	var schedules []models.Schedule
	db.Where("aktif = ?", true).Find(&schedules)

	for _, s := range schedules {
		if s.Jam == currentHour && s.Menit == currentMin {
			// Klaim atomik: hanya 1 proses yang menang per jadwal per hari
			res := db.Model(&models.Schedule{}).
				Where("id = ? AND (last_run IS NULL OR last_run <> ?)", s.ID, todayStr).
				Update("last_run", todayStr)
			if res.Error != nil || res.RowsAffected == 0 {
				continue
			}
			// Fuzzy start: tunda acak 0-90 detik agar tidak jalan tepat
			// di detik :00 setiap hari (jadwal cron-tepat = ciri bot).
			delay := time.Duration(rand.Intn(91)) * time.Second
			log.Printf("[SCHEDULER] auto-scrape %s (%02d:%02d WITA, mode=%s) mulai dalam %ds",
				s.Label, s.Jam, s.Menit, s.Mode, int(delay.Seconds()))
			go func(sched models.Schedule, d time.Duration) {
				time.Sleep(d)
				autoScrapeByMode(db, sched.Mode)
				// Setelah batch dibuat → tunggu habis lalu kirim rekap gambar
				// Menggunakan per-schedule config (telegram_enabled, wa_enabled, wa_group)
				recap.AfterScheduleBatch(db, sched.ID, sched.Label, todayStr, &sched)
			}(s, delay)
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
	db.Where("aktif = ? AND password_pusaka != ''", true).Find(&pegawai)

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
