package handler

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type DashboardHandler struct {
	DB *gorm.DB
}

// Simple in-memory cache with TTL
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type dashboardCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

func newDashboardCache(ttl time.Duration) *dashboardCache {
	c := &dashboardCache{items: make(map[string]cacheEntry), ttl: ttl}
	// Cleanup expired entries every minute
	go func() {
		for range time.Tick(time.Minute) {
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.items {
				if now.After(v.expiresAt) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}()
	return c
}

func (c *dashboardCache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.data, true
}

func (c *dashboardCache) set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheEntry{data: data, expiresAt: time.Now().Add(c.ttl)}
}

// Cache for dashboard data (30 second TTL — data changes when scrape runs)
var dashCache = newDashboardCache(30 * time.Second)

// GET /api/dashboard?tanggal=YYYY-MM-DD
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	tanggal := c.DefaultQuery("tanggal", time.Now().UTC().Add(8*time.Hour).Format("2006-01-02"))

	// Check cache
	instStr, _ := instansiID.(string)
	cacheKey := "daily:" + tanggal + ":" + role.(string) + ":" + instStr
	if cached, ok := dashCache.get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	q := h.DB
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("instansi_id = ? AND tanggal = ?", instansiID, tanggal)
	} else {
		q = h.DB.Where("tanggal = ?", tanggal)
	}

	var absensi []models.Absensi
	q.Order("nama").Find(&absensi)

	// Query rekap TERPISAH (fresh scope)
	rekapQ := h.DB
	if role != "superadmin" && instansiID != "" {
		rekapQ = h.DB.Where("instansi_id = ? AND tanggal = ?", instansiID, tanggal)
	} else {
		rekapQ = h.DB.Where("tanggal = ?", tanggal)
	}

	var rekap models.RekapHarian
	rekapQ.Model(&models.Absensi{}).
		Select(`
			COUNT(*) as total,
			SUM(CASE WHEN jam_masuk IS NOT NULL AND jam_masuk != '-' AND jam_masuk != '' THEN 1 ELSE 0 END) as hadir,
			SUM(CASE WHEN status = 'Terlambat' OR status = 'Telat Ringan' THEN 1 ELSE 0 END) as terlambat,
			SUM(CASE WHEN (jam_masuk IS NULL OR jam_masuk = '-' OR jam_masuk = '') AND (jam_pulang IS NULL OR jam_pulang = '-' OR jam_pulang = '') THEN 1 ELSE 0 END) as tidak_hadir,
			SUM(CASE WHEN status = 'Belum Masuk' THEN 1 ELSE 0 END) as belum_masuk,
			SUM(CASE WHEN status = 'Belum Pulang' THEN 1 ELSE 0 END) as belum_pulang
		`).Scan(&rekap)

	resp := models.ApiResponse{
		Success: true,
		Data: gin.H{
			"tanggal": tanggal,
			"absensi": absensi,
			"rekap":   rekap,
		},
	}

	// Store in cache
	dashCache.set(cacheKey, resp)

	// Set no-cache headers for API response (browser should not cache dynamic data)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.JSON(http.StatusOK, resp)
}

// GET /api/dashboard/bulan?bulan=MM&tahun=YYYY
func (h *DashboardHandler) DashboardBulan(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))

	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

	// Check cache
	instStr, _ := instansiID.(string)
	cacheKey := "bulan:" + pattern + ":" + role.(string) + ":" + instStr
	if cached, ok := dashCache.get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Superadmin lihat semua instansi
	qInstansi := h.DB
	if role != "superadmin" && instansiID != "" {
		qInstansi = h.DB.Where("instansi_id = ?", instansiID)
	}

	// Total pegawai aktif
	var totalPegawai int64
	qInstansi.Model(&models.Pegawai{}).Where("aktif = ?", true).Count(&totalPegawai)

	// Rekap per tanggal
	type RekapTanggal struct {
		Tanggal    string `json:"tanggal"`
		Hadir      int64  `json:"hadir"`
		Terlambat  int64  `json:"terlambat"`
		TidakHadir int64  `json:"tidak_hadir"`
	}
	var rekapHarian []RekapTanggal
	qRekap := h.DB.Model(&models.Absensi{}).
		Select(`
			tanggal as tanggal,
			SUM(CASE WHEN jam_masuk IS NOT NULL AND jam_masuk != '-' AND jam_masuk != '' THEN 1 ELSE 0 END) as hadir,
			SUM(CASE WHEN status = 'Terlambat' THEN 1 ELSE 0 END) as terlambat,
			SUM(CASE WHEN (jam_masuk IS NULL OR jam_masuk = '-' OR jam_masuk = '') AND (jam_pulang IS NULL OR jam_pulang = '-' OR jam_pulang = '') THEN 1 ELSE 0 END) as tidak_hadir
		`)
	if role != "superadmin" && instansiID != "" {
		qRekap = qRekap.Where("instansi_id = ? AND tanggal LIKE ?", instansiID, pattern)
	} else {
		qRekap = qRekap.Where("tanggal LIKE ?", pattern)
	}
	qRekap.Group("tanggal").Order("tanggal").Scan(&rekapHarian)

	// Ringkasan bulan
	var summary struct {
		TotalHari  int64   `json:"total_hari"`
		TotalAbsen int64   `json:"total_absen"`
		RataHadir  float64 `json:"rata_hadir"`
	}
	summary.TotalHari = int64(len(rekapHarian))
	var totalHadir int64
	for _, r := range rekapHarian {
		totalHadir += r.Hadir
	}
	if summary.TotalHari > 0 && totalPegawai > 0 {
		summary.RataHadir = float64(totalHadir) / float64(summary.TotalHari) / float64(totalPegawai) * 100
	}

	resp := models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":        bulan,
			"tahun":        tahun,
			"total_pegawai": totalPegawai,
			"summary":      summary,
			"rekap_harian": rekapHarian,
		},
	}

	dashCache.set(cacheKey, resp)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.JSON(http.StatusOK, resp)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// GET /api/dashboard/bulan-pegawai?bulan=MM&tahun=YYYY
// FIXED: Single aggregate query instead of N+1 (was: 1 + N queries, now: 2 queries)
func (h *DashboardHandler) DashboardBulanPegawai(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))
	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

	// Check cache
	instStr, _ := instansiID.(string)
	cacheKey := "bulan-peg:" + pattern + ":" + role.(string) + ":" + instStr
	if cached, ok := dashCache.get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Hitung hari kerja di bulan ini
	var hariKerja int64
	qHK := h.DB.Model(&models.Absensi{})
	if role != "superadmin" && instansiID != "" {
		qHK = qHK.Where("instansi_id = ? AND tanggal LIKE ?", instansiID, pattern)
	} else {
		qHK = qHK.Where("tanggal LIKE ?", pattern)
	}
	qHK.Distinct("tanggal").Count(&hariKerja)

	// FIXED: Single aggregate query — GROUP BY n_ip, nama instead of N+1 loop
	type PegRekap struct {
		NIP          string  `gorm:"column:n_ip" json:"nip"`
		Nama         string  `gorm:"column:nama" json:"nama"`
		Hadir        int64   `gorm:"column:hadir" json:"hadir"`
		Terlambat    int64   `gorm:"column:terlambat" json:"terlambat"`
		TidakHadir   int64   `gorm:"column:tidak_hadir" json:"tidak_hadir"`
		BelumMasuk   int64   `gorm:"column:belum_masuk" json:"belum_masuk"`
		BelumPulang  int64   `gorm:"column:belum_pulang" json:"belum_pulang"`
	}

	var rekapRows []PegRekap
	rawQuery := `
		SELECT
			p.n_ip,
			p.nama,
			COALESCE(SUM(CASE WHEN a.jam_masuk IS NOT NULL AND a.jam_masuk != '-' AND a.jam_masuk != '' THEN 1 ELSE 0 END), 0) as hadir,
			COALESCE(SUM(CASE WHEN a.status = 'Terlambat' OR a.status = 'Telat Ringan' THEN 1 ELSE 0 END), 0) as terlambat,
			COALESCE(SUM(CASE WHEN (a.jam_masuk IS NULL OR a.jam_masuk = '-' OR a.jam_masuk = '') AND (a.jam_pulang IS NULL OR a.jam_pulang = '-' OR a.jam_pulang = '') THEN 1 ELSE 0 END), 0) as tidak_hadir,
			COALESCE(SUM(CASE WHEN a.status = 'Belum Masuk' THEN 1 ELSE 0 END), 0) as belum_masuk,
			COALESCE(SUM(CASE WHEN a.status = 'Belum Pulang' THEN 1 ELSE 0 END), 0) as belum_pulang
		FROM pegawais p
		LEFT JOIN absensis a ON a.n_ip = p.n_ip AND a.tanggal LIKE ?` + ""

	if role != "superadmin" && instansiID != "" {
		rawQuery += ` AND a.instansi_id = ?`
		rawQuery += `
		WHERE p.aktif = ? AND p.instansi_id = ?
		GROUP BY p.n_ip, p.nama
		ORDER BY p.nama`
		h.DB.Raw(rawQuery, pattern, instansiID, true, instansiID).Scan(&rekapRows)
	} else {
		rawQuery += `
		WHERE p.aktif = ?
		GROUP BY p.n_ip, p.nama
		ORDER BY p.nama`
		h.DB.Raw(rawQuery, pattern, true).Scan(&rekapRows)
	}

	// Build result with percentage
	result := make([]gin.H, 0, len(rekapRows))
	for _, r := range rekapRows {
		persen := 0.0
		if hariKerja > 0 {
			persen = float64(r.Hadir) / float64(hariKerja) * 100
		}
		result = append(result, gin.H{
			"nip":          r.NIP,
			"nama":         r.Nama,
			"hadir":        r.Hadir,
			"terlambat":    r.Terlambat,
			"tidak_hadir":  r.TidakHadir,
			"belum_masuk":  r.BelumMasuk,
			"belum_pulang": r.BelumPulang,
			"persen":       persen,
		})
	}

	resp := models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":         bulan,
			"tahun":         tahun,
			"hari_kerja":    hariKerja,
			"rekap_pegawai": result,
		},
	}

	dashCache.set(cacheKey, resp)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.JSON(http.StatusOK, resp)
}

// GET /api/dashboard/bulan-detail?bulan=MM&tahun=YYYY
func (h *DashboardHandler) DashboardBulanDetail(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))
	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

	// Check cache
	instStr, _ := instansiID.(string)
	cacheKey := "bulan-detail:" + pattern + ":" + role.(string) + ":" + instStr
	if cached, ok := dashCache.get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	q := h.DB.Model(&models.Absensi{}).
		Select("tanggal, nama, jam_masuk, jam_pulang, status").
		Order("tanggal DESC, nama")
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ? AND tanggal LIKE ?", instansiID, pattern)
	} else {
		q = q.Where("tanggal LIKE ?", pattern)
	}

	var rows []models.Absensi
	q.Scan(&rows)

	resp := models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":  bulan,
			"tahun":  tahun,
			"detail": rows,
		},
	}

	dashCache.set(cacheKey, resp)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.JSON(http.StatusOK, resp)
}
