package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// queryScrapeStatus — stats + 500 job terbaru (dipakai Status & Stream).
// PENTING: tiap query pakai Session NewDB agar klausa SELECT(SUM) tidak
// bocor ke query berikutnya (Postgres menolak ORDER BY di atas agregat;
// SQLite meloloskan — bug sekelas dashboard kemarin).
func queryScrapeStatus(db *gorm.DB, instansiID, role string) (models.JobStats, []models.Job, error) {
	newScope := func() *gorm.DB {
		s := db.Session(&gorm.Session{NewDB: true})
		if role != "superadmin" && instansiID != "" {
			s = s.Where("instansi_id = ?", instansiID)
		}
		return s
	}

	var stats models.JobStats
	if err := newScope().Model(&models.Job{}).
		Select(`
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
			SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) as done,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled
		`).Scan(&stats).Error; err != nil {
		return stats, nil, err
	}

	// Urutan kategori: running paling atas, pending tengah, done bawah,
	// failed + cancelled paling bawah. Dalam kategori: terbaru dulu.
	// Limit 100 agar ringan (daftar 500 baris terlalu berat untuk cloud).
	var jobs []models.Job
	if err := newScope().Order(`
		CASE status
			WHEN 'running' THEN 0
			WHEN 'pending' THEN 1
			WHEN 'done' THEN 2
			WHEN 'failed' THEN 3
			ELSE 4
		END, created_at DESC`).Limit(100).Find(&jobs).Error; err != nil {
		return stats, nil, err
	}
	return stats, jobs, nil
}

// statusSignature — sidik jari ringan untuk deteksi perubahan antrean.
// Stats selalu dikirim tiap tick (kecil); daftar job (500 baris) hanya
// dikirim ulang kalau signature berubah → hemat bandwidth ke DB cloud.
func statusSignature(stats models.JobStats, jobs []models.Job) string {
	active := 0
	progSum := 0
	var latest string
	var latestProg string
	for _, j := range jobs {
		if j.Status == "pending" || j.Status == "running" {
			active++
			progSum += j.Progress
			// Sertakan label langkah terakhir agar tiap step memicu push
			if s := j.StepLabel; s > latestProg {
				latestProg = s
			}
		}
		ts := j.CreatedAt.Format("060102150405")
		if ca := j.ClaimedAt; ca != nil {
			if s := ca.Format("060102150405"); s > ts {
				ts = s
			}
		}
		if ca := j.CompletedAt; ca != nil {
			if s := ca.Format("060102150405"); s > ts {
				ts = s
			}
		}
		if ts > latest {
			latest = ts
		}
	}
	return fmt.Sprintf("%d-%d-%d-%d|%d|%d|%s|%d|%s",
		stats.Pending, stats.Running, stats.Done, stats.Failed,
		len(jobs), active, latest, progSum, latestProg)
}

// GET /api/scrape/stream — Server-Sent Events untuk status antrean.
// Menggantikan polling 500ms: 1 koneksi per tab, server tick tiap 2 detik,
// stats dikirim tiap tick, daftar job hanya saat berubah.
func (h *ScrapeHandler) StreamStatus(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	instID, _ := instansiID.(string)
	roleStr, _ := role.(string)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	send := func(event string, v interface{}) bool {
		b, err := json.Marshal(v)
		if err != nil {
			return true
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, b); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()
	lastSig := ""

	// Kirim snapshot awal langsung (tanpa tunggu tick pertama)
	if stats, jobs, err := queryScrapeStatus(h.DB, instID, roleStr); err == nil {
		if !send("stats", stats) {
			return
		}
		if !send("jobs", jobs) {
			return
		}
		lastSig = statusSignature(stats, jobs)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, jobs, err := queryScrapeStatus(h.DB, instID, roleStr)
			if err != nil {
				continue
			}
			if !send("stats", stats) {
				return
			}
			if sig := statusSignature(stats, jobs); sig != lastSig {
				lastSig = sig
				if !send("jobs", jobs) {
					return
				}
			} else if !send("ping", gin.H{"t": time.Now().Unix()}) {
				return
			}
		}
	}
}
