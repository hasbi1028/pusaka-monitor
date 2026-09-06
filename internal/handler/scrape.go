package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hasbiawal/pusaka-monitor/internal/crypto"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
	"github.com/hasbiawal/pusaka-monitor/internal/scraper"
)

type ScrapeHandler struct {
	DB *gorm.DB
}

// POST /api/scrape  (mode: all | belum_masuk | belum_pulang)
func (h *ScrapeHandler) CreateJobs(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	mode := c.DefaultQuery("mode", "all") // all | belum_masuk | belum_pulang
	today := todayWITA()

	q := h.DB.Where("aktif = ? AND password_pusaka != ''", true)
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("instansi_id = ? AND aktif = ? AND password_pusaka != ''", instansiID, true)
	}

	var pegawaiList []models.Pegawai
	q.Find(&pegawaiList)

	created := 0
	for _, p := range pegawaiList {
		instID := instansiID.(string)
		if role == "superadmin" {
			instID = p.InstansiID
		}
		// Cek absensi hari ini
		var existing models.Absensi
		hasAbsen := h.DB.Where("n_ip = ? AND instansi_id = ? AND tanggal = ?", p.NIP, instID, today).First(&existing).Error == nil

		switch mode {
			case "belum_masuk":
				// Hanya yg BELUM punya absensi hari ini (atau jam_masuk kosong)
				if hasAbsen && existing.JamMasuk != "" && existing.JamMasuk != "-" {
					continue
				}
			case "belum_pulang":
				// Hanya yg SUDAH masuk tapi BELUM pulang
				if !(hasAbsen && existing.JamMasuk != "" && existing.JamMasuk != "-" && (existing.JamPulang == "" || existing.JamPulang == "-")) {
					continue
				}
			default: // all — scrape SEMUA pegawai, tanpa skip
				// Tidak ada filter: semua pegawai dibuat job-nya
			}

		job := models.Job{
			ID:           uuid.New().String(),
			InstansiID:   instID,
			EmployeeID:   p.NIP,
			EmployeeName: p.Nama,
			Status:       "pending",
			MaxAttempts:  3,
		}
		if h.DB.Create(&job).Error == nil {
			created++
		}
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job dibuat",
		Data:    gin.H{"created": created},
	})
}

// todayWITA — tanggal hari ini (WITA = UTC+8) format YYYY-MM-DD
func todayWITA() string {
	return time.Now().UTC().Add(8 * time.Hour).Format("2006-01-02")
}

// isLengkap — cek apakah absensi sudah masuk DAN pulang
func isLengkap(masuk, pulang string) bool {
	empty := func(s string) bool { return s == "" || s == "-" }
	return !empty(masuk) && !empty(pulang)
}

// GET /api/scrape/status
func (h *ScrapeHandler) Status(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	instID, _ := instansiID.(string)
	roleStr, _ := role.(string)

	stats, jobs, err := queryScrapeStatus(h.DB, instID, roleStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Success: false, Error: "Gagal memuat status"})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"stats": stats, "jobs": jobs},
	})
}

// POST /api/scrape/retry
func (h *ScrapeHandler) RetryFailed(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")

	result := h.DB.Model(&models.Job{}).
		Where("instansi_id = ? AND status = 'failed'", instansiID).
		Updates(map[string]interface{}{
			"status":      "pending",
			"attempts":    0,
			"error":       "",
			"worker_id":   "",
			"claimed_at":  gorm.Expr("NULL"),
		})

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job di-retry",
		Data:    gin.H{"reset": result.RowsAffected},
	})
}

// POST /api/scrape/cancel-all — batalkan semua job pending & running
func (h *ScrapeHandler) CancelAll(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")

	result := h.DB.Model(&models.Job{}).
		Where("instansi_id = ? AND status IN ('pending','running')", instansiID).
		Updates(map[string]interface{}{
			"status":      "cancelled",
			"worker_id":   "",
			"claimed_at":  gorm.Expr("NULL"),
		})

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Semua job dibatalkan",
		Data:    gin.H{"cancelled": result.RowsAffected},
	})
}

// POST /api/scrape/pegawai/:nip — scrape 1 pegawai saja
func (h *ScrapeHandler) ScrapeOne(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	nip := c.Param("nip")
	force := c.Query("force") == "1"

	// Cari pegawai (superadmin lihat semua instansi)
	var pegawai models.Pegawai
	q := h.DB.Where("n_ip = ?", nip)
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("n_ip = ? AND instansi_id = ?", nip, instansiID)
	}
	if q.First(&pegawai).Error != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Success: false, Error: "Pegawai tidak ditemukan"})
		return
	}

	// Skip kalau sudah lengkap (kecuali force)
	if !force {
		var existing models.Absensi
		if h.DB.Where("n_ip = ? AND instansi_id = ? AND tanggal = ?", nip, pegawai.InstansiID, todayWITA()).First(&existing).Error == nil {
			if isLengkap(existing.JamMasuk, existing.JamPulang) {
				c.JSON(http.StatusOK, models.ApiResponse{
					Success: true,
					Message: "Sudah lengkap (masuk+pulang) — skip. Pakai ?force=1 untuk scrape paksa",
					Data:    gin.H{"skipped": true, "status": existing.Status},
				})
				return
			}
		}
	}

	job := models.Job{
		ID:           uuid.New().String(),
		InstansiID:   pegawai.InstansiID,
		EmployeeID:   pegawai.NIP,
		EmployeeName: pegawai.Nama,
		Status:       "pending",
		MaxAttempts:  3,
	}
	if h.DB.Create(&job).Error != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Success: false, Error: "Gagal membuat job"})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job 1 pegawai dibuat: " + pegawai.Nama,
		Data:    gin.H{"job_id": job.ID, "nip": nip},
	})
}

// POST /api/scrape/:id/cancel — batalkan 1 job (by ID)
func (h *ScrapeHandler) CancelOne(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	id := c.Param("id")

	result := h.DB.Model(&models.Job{}).
		Where("id = ? AND instansi_id = ? AND status IN ('pending','running')", id, instansiID).
		Updates(map[string]interface{}{
			"status":      "cancelled",
			"worker_id":   "",
			"claimed_at":  gorm.Expr("NULL"),
		})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ApiResponse{Success: false, Error: "Job tidak ditemukan atau sudah selesai"})
		return
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Job dibatalkan",
		Data:    gin.H{"id": id},
	})
}

// Background worker — pool dengan N concurrent worker
var claimMu sync.Mutex

// claimNext — ambil 1 job pending secara atomik (pakai mutex agar 8 worker tidak double-claim)
func claimNext(db *gorm.DB) (models.Job, bool) {
	claimMu.Lock()
	defer claimMu.Unlock()

	var job models.Job
	if db.Where("status = ?", "pending").Order("created_at ASC").First(&job).Error != nil {
		return job, false
	}
	now := time.Now()
	db.Model(&job).Updates(map[string]interface{}{
		"status":      "running",
		"worker_id":   "worker-localhost",
		"claimed_at":  &now,
		"attempts":    job.Attempts + 1,
	})
	return job, true
}

func processJob(db *gorm.DB, job models.Job) {
	now := time.Now()
	today := todayWITA()

	// Get password (cari by NIP; instansi_id bisa kosong untuk superadmin)
	var pegawai models.Pegawai
	q := db.Where("n_ip = ?", job.EmployeeID)
	if job.InstansiID != "" {
		q = db.Where("n_ip = ? AND instansi_id = ?", job.EmployeeID, job.InstansiID)
	}
	if q.First(&pegawai).Error != nil {
		db.Model(&job).Updates(map[string]interface{}{
			"status":       "failed",
			"error":        "Pegawai tidak ditemukan",
			"completed_at": &now,
		})
		return
	}

	// SKILL: jika sudah lengkap (masuk + pulang) → skip scrape
	var existing models.Absensi
	if db.Where("n_ip = ? AND instansi_id = ? AND tanggal = ?", job.EmployeeID, pegawai.InstansiID, today).First(&existing).Error == nil {
		if isLengkap(existing.JamMasuk, existing.JamPulang) {
			db.Model(&job).Updates(map[string]interface{}{
				"status":       "done",
				"tanggal":      today,
				"jam_masuk":    existing.JamMasuk,
				"jam_pulang":   existing.JamPulang,
				"completed_at": &now,
			})
			return
		}
	}

	// Stagger: tiap job mulai dengan jeda acak 0-90 detik agar 8 worker
	// tidak login ke Pusaka di detik yang sama (thundering herd = ciri bot).
	// Progres 0 + label "Antre giliran" supaya terlihat di SSE stream.
	db.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
		"progress":    0,
		"total_steps": scraper.TotalSteps,
		"step_label":  "Antre giliran",
	})
	time.Sleep(time.Duration(rand.Intn(91)) * time.Second)

	// Decrypt password Pusaka sebelum scrape
	pwd := pegawai.PasswordPusaka
	if pwd != "" {
		if decrypted, err := crypto.Decrypt(pwd); err == nil {
			pwd = decrypted
		}
	}

	// Scrape via HTTP API client (Pusaka v3) — progres ditulis ke DB tiap langkah
	client := scraper.NewClient(pegawai.NIP, pegawai.Nama, pwd)
	client.OnProgress = func(step int, label string) {
		db.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"progress":    step,
			"total_steps": scraper.TotalSteps,
			"step_label":  label,
		})
	}
	result := client.ScrapeToday()

	log.Printf("[SCRAPER] %s (%s) → success=%v, masuk=%s, pulang=%s, err=%s",
		pegawai.Nama, job.EmployeeID, result.Success, result.JamMasuk, result.JamPulang, result.Error)

	// Jika data belum ada di Pusaka → cek dulu apakah hari libur
	if !result.Success {
		status := "Belum Masuk"
		// Cek hari libur (mingguan instansi + tanggal merah)
		if IsLibur(db, pegawai.InstansiID, today) {
			status = "Libur"
		}
		// Cek cuti
		var cutiCount int64
		db.Model(&models.Cuti{}).Where("n_ip = ? AND tanggal_mulai <= ? AND tanggal_akhir >= ?",
			job.EmployeeID, today, today).Count(&cutiCount)
		if cutiCount > 0 {
			status = "Cuti"
		}

		record := models.Absensi{
			ID:         uuid.New().String(),
			InstansiID: pegawai.InstansiID,
			NIP:        job.EmployeeID,
			Nama:       pegawai.Nama,
			Tanggal:    today,
			JamMasuk:   "-",
			JamPulang:  "-",
			Status:     status,
			ScrapedAt:  now,
		}
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "n_ip"}, {Name: "tanggal"}, {Name: "instansi_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"nama", "scraped_at"}),
		}).Create(&record)

		db.Model(&job).Updates(map[string]interface{}{
			"status":       "done",
			"tanggal":      today,
			"jam_masuk":    "-",
			"jam_pulang":   "-",
			"error":        result.Error,
			"progress":     scraper.TotalSteps,
			"total_steps":  scraper.TotalSteps,
			"step_label":   "Selesai",
			"completed_at": &now,
		})
		return
	}

	// Ambil jam kerja dari instansi
	var instansi models.Instansi
	db.First(&instansi, "id = ?", pegawai.InstansiID)

	// Pilih jam kerja berdasarkan mode ramadan
	jamMasukStd := instansi.JamMasuk
	toleransi := instansi.Toleransi
	if instansi.ModeRamadan {
		jamMasukStd = instansi.JamMasukRam
		toleransi = instansi.ToleransiRam
	}

	// Simpan ke tabel absensi (UPSERT natif via unique index)
	db.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
		"progress":    9,
		"total_steps": scraper.TotalSteps,
		"step_label":  "Simpan",
	})
	absStatus := scraper.DetermineStatus(result.JamMasuk, result.JamPulang, jamMasukStd, toleransi)
	record := models.Absensi{
		ID:         uuid.New().String(),
		InstansiID: pegawai.InstansiID,
		NIP:        job.EmployeeID,
		Nama:       pegawai.Nama,
		Tanggal:    result.Tanggal,
		JamMasuk:   result.JamMasuk,
		JamPulang:  result.JamPulang,
		Status:     absStatus,
		ScrapedAt:  now,
	}
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "n_ip"}, {Name: "tanggal"}, {Name: "instansi_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"jam_masuk", "jam_pulang", "status", "nama", "scraped_at"}),
	}).Create(&record)

	db.Model(&job).Updates(map[string]interface{}{
		"status":       "done",
		"tanggal":      result.Tanggal,
		"jam_masuk":    result.JamMasuk,
		"jam_pulang":   result.JamPulang,
		"progress":     scraper.TotalSteps,
		"total_steps":  scraper.TotalSteps,
		"step_label":   "Selesai",
		"completed_at": &now,
	})
}

// Background worker — pool statis MAX_WORKERS, concurrency aktif diatur via atomic (bisa diubah superadmin)
const MaxWorkers = 100

var (
	targetConcurrency int32 = 8 // nilai default, di-override dari DB saat start
)

// SetConcurrency — ubah jumlah worker aktif (dipanggil dari endpoint superadmin)
func SetConcurrency(n int) {
	if n < 1 {
		n = 1
	}
	if n > MaxWorkers {
		n = MaxWorkers
	}
	atomic.StoreInt32(&targetConcurrency, int32(n))
}

func getConcurrency() int32 {
	return atomic.LoadInt32(&targetConcurrency)
}

// POST /api/admin/concurrency  body: {"concurrency": N}
func (h *ScrapeHandler) SetConcurrencyHandler(c *gin.Context) {
	var body struct {
		Concurrency int `json:"concurrency"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Concurrency < 1 {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Success: false, Error: "concurrency harus >= 1"})
		return
	}
	SetConcurrency(body.Concurrency)
	// Simpan ke DB
	h.DB.Save(&models.Setting{Key: "worker_concurrency", Value: fmt.Sprintf("%d", body.Concurrency), InstansiID: "global", UpdatedAt: time.Now()})
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: gin.H{"concurrency": body.Concurrency}})
}

// GET /api/admin/concurrency
func (h *ScrapeHandler) GetConcurrencyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: gin.H{"concurrency": getConcurrency()}})
}
func StartWorker(db *gorm.DB) {
	// Load dari DB (settings key worker_concurrency)
	if v, err := getSettingInt(db, "worker_concurrency"); err == nil && v > 0 {
		SetConcurrency(v)
	}
	for i := 0; i < MaxWorkers; i++ {
		go func(workerID int) {
			done := 0
			nextBreak := 5 + rand.Intn(4) // istirahat panjang tiap 5-8 job
			for {
				if int32(workerID) >= effectiveConcurrency(db) {
					// Idle (di atas target) — tidur lama, nilai pending di-cache 10 dtk
					time.Sleep(10 * time.Second)
					continue
				}
				job, ok := claimNext(db)
				if !ok {
					time.Sleep(10 * time.Second)
					continue
				}
				processJob(db, job)
				done++
				if done >= nextBreak {
					// Istirahat panjang 1-3 menit (pola manusia)
					done = 0
					nextBreak = 5 + rand.Intn(4)
					time.Sleep(time.Duration(60+rand.Intn(121)) * time.Second)
				} else {
					time.Sleep(jedaNatural())
				}
			}
		}(i)
	}
}

// jedaNatural — jeda antar job pola manusia: 80% pendek (5-12 dtk),
// 20% panjang (20-45 dtk). Uniform murni mudah dikenali sebagai bot.
func jedaNatural() time.Duration {
	if rand.Intn(100) < 20 {
		return time.Duration(20+rand.Intn(26)) * time.Second
	}
	return time.Duration(5000+rand.Intn(7000)) * time.Millisecond
}

// Profil concurrency dinamis — hindari beban "kotak sempurna":
// ramp-up 3 menit dari 3 worker ke target, taper ke 3 worker saat sisa < 10.
var batchMu sync.Mutex
var batchStart time.Time
var batchWasIdle = true

// Cache hitungan pending — 100 worker idle JANGAN query DB tiap 2 dtk.
// Satu nilai bersama, refresh maksimal tiap 10 dtk.
var pendingCacheMu sync.Mutex
var pendingCacheCount int64
var pendingCacheAt time.Time

func cachedPending(db *gorm.DB) int64 {
	pendingCacheMu.Lock()
	defer pendingCacheMu.Unlock()
	if time.Since(pendingCacheAt) < 10*time.Second {
		return pendingCacheCount
	}
	var pending int64
	db.Model(&models.Job{}).Where("status = ?", "pending").Count(&pending)
	pendingCacheCount = pending
	pendingCacheAt = time.Now()
	return pending
}

func effectiveConcurrency(db *gorm.DB) int32 {
	target := getConcurrency()
	pending := cachedPending(db)
	if pending == 0 {
		batchMu.Lock()
		batchWasIdle = true
		batchMu.Unlock()
		return target
	}
	batchMu.Lock()
	if batchWasIdle {
		batchStart = time.Now()
		batchWasIdle = false
	}
	elapsed := time.Since(batchStart)
	batchMu.Unlock()

	eff := target
	if elapsed < 3*time.Minute && target > 3 {
		eff = 3 + int32(float64(target-3)*float64(elapsed)/float64(3*time.Minute))
	}
	if pending < 10 && eff > 3 {
		eff = 3
	}
	if eff < 1 {
		eff = 1
	}
	return eff
}

// getSettingInt — baca setting dari DB
func getSettingInt(db *gorm.DB, key string) (int, error) {
	var s models.Setting
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		return 0, err
	}
	var n int
	_, err := fmt.Sscanf(s.Value, "%d", &n)
	return n, err
}
