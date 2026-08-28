package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type DashboardHandler struct {
	DB *gorm.DB
}

// GET /api/dashboard?tanggal=YYYY-MM-DD
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	tanggal := c.DefaultQuery("tanggal", time.Now().UTC().Add(8*time.Hour).Format("2006-01-02"))

	q := h.DB
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("instansi_id = ? AND tanggal = ?", instansiID, tanggal)
	} else {
		q = h.DB.Where("tanggal = ?", tanggal)
	}

	var absensi []models.Absensi
	q.Order("nama").Find(&absensi)

	var rekap models.RekapHarian
	q.Model(&models.Absensi{}).
		Select(`
			COUNT(*) as total,
			SUM(CASE WHEN jam_masuk IS NOT NULL AND jam_masuk != '-' AND jam_masuk != '' THEN 1 ELSE 0 END) as hadir,
			SUM(CASE WHEN status = 'Terlambat' OR status = 'Telat Ringan' THEN 1 ELSE 0 END) as terlambat,
			SUM(CASE WHEN (jam_masuk IS NULL OR jam_masuk = '-' OR jam_masuk = '') AND (jam_pulang IS NULL OR jam_pulang = '-' OR jam_pulang = '') THEN 1 ELSE 0 END) as tidak_hadir,
			SUM(CASE WHEN status = 'Belum Masuk' THEN 1 ELSE 0 END) as belum_masuk,
			SUM(CASE WHEN status = 'Belum Pulang' THEN 1 ELSE 0 END) as belum_pulang
		`).Scan(&rekap)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"tanggal": tanggal,
			"absensi": absensi,
			"rekap":   rekap,
		},
	})
}

// GET /api/dashboard/bulan?bulan=MM&tahun=YYYY
// Rekap per bulan: total hari, rata-rata hadir, daftar tanggal tidak hadir per pegawai
func (h *DashboardHandler) DashboardBulan(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))

	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

	// Superadmin lihat semua instansi
	qInstansi := h.DB
	if role != "superadmin" && instansiID != "" {
		qInstansi = h.DB.Where("instansi_id = ?", instansiID)
	}

	// Total pegawai aktif
	var totalPegawai int64
	qInstansi.Model(&models.Pegawai{}).Where("aktif = 1").Count(&totalPegawai)

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
		TotalHari  int64 `json:"total_hari"`
		TotalAbsen int64 `json:"total_absen"`
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

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":       bulan,
			"tahun":       tahun,
			"total_pegawai": totalPegawai,
			"summary":     summary,
			"rekap_harian": rekapHarian,
		},
	})
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// GET /api/dashboard/bulan-pegawai?bulan=MM&tahun=YYYY
// Rekap per pegawai untuk 1 bulan: hadir, telat, alfa, % kehadiran.
func (h *DashboardHandler) DashboardBulanPegawai(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))
	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

	// Ambil semua pegawai aktif (scope instansi)
	type PegRow struct {
		NIP  string `gorm:"column:n_ip"`
		Nama string `gorm:"column:nama"`
	}
	var pegawai []PegRow
	qPeg := h.DB.Model(&models.Pegawai{}).Select("n_ip, nama")
	if role != "superadmin" && instansiID != "" {
		qPeg = qPeg.Where("instansi_id = ? AND aktif = 1", instansiID)
	} else {
		qPeg = qPeg.Where("aktif = 1")
	}
	qPeg.Order("nama").Scan(&pegawai)
	// Debug: pastikan NIP terisi
	if len(pegawai) > 0 && pegawai[0].NIP == "" {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Success: false, Error: "NIP kosong (bug select)"})
		return
	}

	// Hitung hari kerja di bulan ini (ada data absensi)
	var hariKerja int64
	qHK := h.DB.Model(&models.Absensi{})
	if role != "superadmin" && instansiID != "" {
		qHK = qHK.Where("instansi_id = ? AND tanggal LIKE ?", instansiID, pattern)
	} else {
		qHK = qHK.Where("tanggal LIKE ?", pattern)
	}
	qHK.Distinct("tanggal").Count(&hariKerja)

	result := []gin.H{}
	for _, p := range pegawai {
		var r struct {
			Hadir       int64 `gorm:"column:hadir" json:"hadir"`
			Terlambat   int64 `gorm:"column:terlambat" json:"terlambat"`
			TidakHadir  int64 `gorm:"column:tidak_hadir" json:"tidak_hadir"`
			BelumMasuk  int64 `gorm:"column:belum_masuk" json:"belum_masuk"`
			BelumPulang int64 `gorm:"column:belum_pulang" json:"belum_pulang"`
		}
		cond := "n_ip = ? AND tanggal LIKE ?"
		args := []interface{}{p.NIP, pattern}
		if role != "superadmin" && instansiID != "" {
			cond = "n_ip = ? AND instansi_id = ? AND tanggal LIKE ?"
			args = []interface{}{p.NIP, instansiID, pattern}
		}
		h.DB.Raw(`
			SELECT
				SUM(CASE WHEN jam_masuk IS NOT NULL AND jam_masuk != '-' AND jam_masuk != '' THEN 1 ELSE 0 END) as hadir,
				SUM(CASE WHEN status = 'Terlambat' OR status = 'Telat Ringan' THEN 1 ELSE 0 END) as terlambat,
				SUM(CASE WHEN (jam_masuk IS NULL OR jam_masuk = '-' OR jam_masuk = '') AND (jam_pulang IS NULL OR jam_pulang = '-' OR jam_pulang = '') THEN 1 ELSE 0 END) as tidak_hadir,
				SUM(CASE WHEN status = 'Belum Masuk' THEN 1 ELSE 0 END) as belum_masuk,
				SUM(CASE WHEN status = 'Belum Pulang' THEN 1 ELSE 0 END) as belum_pulang
			FROM absensis WHERE `+cond, args...).Scan(&r)

		persen := 0.0
		if hariKerja > 0 {
			persen = float64(r.Hadir) / float64(hariKerja) * 100
		}
		result = append(result, gin.H{
			"nip":          p.NIP,
			"nama":         p.Nama,
			"hadir":        r.Hadir,
			"terlambat":    r.Terlambat,
			"tidak_hadir":  r.TidakHadir,
			"belum_masuk":  r.BelumMasuk,
			"belum_pulang": r.BelumPulang,
			"persen":       persen,
		})
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":        bulan,
			"tahun":        tahun,
			"hari_kerja":   hariKerja,
			"rekap_pegawai": result,
		},
	})
}

// GET /api/dashboard/bulan-detail?bulan=MM&tahun=YYYY
// Detail flat: 1 baris = 1 presensi (tanggal, nama, masuk, pulang, status).
func (h *DashboardHandler) DashboardBulanDetail(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	now := time.Now().UTC().Add(8 * time.Hour)
	bulan, _ := strconv.Atoi(c.DefaultQuery("bulan", strconv.Itoa(int(now.Month()))))
	tahun, _ := strconv.Atoi(c.DefaultQuery("tahun", strconv.Itoa(now.Year())))
	pattern := strconv.Itoa(tahun) + "-" + pad2(bulan) + "-%"

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

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"bulan":  bulan,
			"tahun":  tahun,
			"detail": rows,
		},
	})
}
