package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// Hari libur mingguan disimpan di settings key "libur_mingguan"
// value: daftar angka hari dipisah koma (0=Minggu..6=Sabtu). Default "0".
func getLiburMingguan(db *gorm.DB, instansiID string) map[int]bool {
	val := ""
	var s models.Setting
	if instansiID != "" {
		if err := db.Where("key = ? AND instansi_id = ?", "libur_mingguan", instansiID).First(&s).Error; err == nil {
			val = s.Value
		}
	}
	if val == "" {
		var g models.Setting
		if err := db.Where("key = ? AND instansi_id = ?", "libur_mingguan", "global").First(&g).Error; err == nil {
			val = g.Value
		}
	}
	if val == "" {
		val = "0" // default: Minggu libur
	}
	out := map[int]bool{}
	for _, p := range strings.Split(val, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && n >= 0 && n <= 6 {
			out[n] = true
		}
	}
	return out
}

// IsLibur — true bila tanggal adalah hari libur mingguan instansi
// ATAU terdaftar di tabel hari_libur (instansi sendiri / global "").
func IsLibur(db *gorm.DB, instansiID, tanggal string) bool {
	t, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return false
	}
	if getLiburMingguan(db, instansiID)[int(t.Weekday())] {
		return true
	}
	var n int64
	db.Model(&models.HariLibur{}).
		Where("tanggal = ? AND (instansi_id = ? OR instansi_id = '')", tanggal, instansiID).
		Count(&n)
	return n > 0
}

// IsCuti — true bila nip sedang cuti pada tanggal (YYYY-MM-DD).
func IsCuti(db *gorm.DB, nip, tanggal string) bool {
	var n int64
	db.Model(&models.Cuti{}).
		Where("n_ip = ? AND tanggal_mulai <= ? AND tanggal_akhir >= ?", nip, tanggal, tanggal).
		Count(&n)
	return n > 0
}

type LiburHandler struct {
	DB *gorm.DB
}

func liburScope(c *gin.Context) (instansiID, role string) {
	if v, ok := c.Get("instansi_id"); ok {
		if s, ok := v.(string); ok {
			instansiID = s
		}
	}
	if v, ok := c.Get("role"); ok {
		if s, ok := v.(string); ok {
			role = s
		}
	}
	return instansiID, role
}

// GET /api/libur?tanggal=YYYY-MM-DD (opsional)
func (h *LiburHandler) Get(c *gin.Context) {
	instansiID, role := liburScope(c)

	mingguan := []int{}
	for d := range getLiburMingguan(h.DB, instansiID) {
		mingguan = append(mingguan, d)
	}

	q := h.DB.Order("tanggal ASC")
	// admin instansi: milik sendiri + global; superadmin: semua
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ? OR instansi_id = ''", instansiID)
	}
	var list []models.HariLibur
	q.Find(&list)

	out := gin.H{"mingguan": mingguan, "tanggal": list}
	if tgl := c.Query("tanggal"); tgl != "" {
		out["is_libur"] = IsLibur(h.DB, instansiID, tgl)
		out["tanggal_cek"] = tgl
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: out})
}

// PUT /api/libur/mingguan  body: {"hari":[0,6]}
func (h *LiburHandler) SetMingguan(c *gin.Context) {
	instansiID, role := liburScope(c)
	key := "global"
	if role != "superadmin" && instansiID != "" {
		key = instansiID
	}
	var req struct {
		Hari []int `json:"hari"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tidak valid"})
		return
	}
	seen := map[int]bool{}
	parts := []string{}
	for _, d := range req.Hari {
		if d < 0 || d > 6 || seen[d] {
			continue
		}
		seen[d] = true
		parts = append(parts, strconv.Itoa(d))
	}
	h.DB.Where("key = ? AND instansi_id = ?", "libur_mingguan", key).
		Assign(models.Setting{Value: strings.Join(parts, ","), UpdatedAt: time.Now()}).
		FirstOrCreate(&models.Setting{Key: "libur_mingguan", InstansiID: key})
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Hari libur mingguan disimpan"})
}

// POST /api/libur/tanggal  body: {"tanggal":"YYYY-MM-DD","keterangan":"..."}
func (h *LiburHandler) AddTanggal(c *gin.Context) {
	instansiID, role := liburScope(c)
	var req struct {
		Tanggal    string `json:"tanggal" binding:"required"`
		Keterangan string `json:"keterangan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Tanggal wajib diisi (YYYY-MM-DD)"})
		return
	}
	if _, err := time.Parse("2006-01-02", req.Tanggal); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tanggal harus YYYY-MM-DD"})
		return
	}
	key := ""
	if role != "superadmin" && instansiID != "" {
		key = instansiID
	}
	rec := models.HariLibur{
		ID:         uuid.New().String(),
		InstansiID: key,
		Tanggal:    req.Tanggal,
		Keterangan: req.Keterangan,
		CreatedAt:  time.Now(),
	}
	if err := h.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal menyimpan"})
		return
	}
	c.JSON(http.StatusCreated, models.ApiResponse{Success: true, Data: rec})
}

// DELETE /api/libur/tanggal/:id
func (h *LiburHandler) DeleteTanggal(c *gin.Context) {
	instansiID, role := liburScope(c)
	id := c.Param("id")
	q := h.DB.Where("id = ?", id)
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ?", instansiID)
	}
	if q.Delete(&models.HariLibur{}).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Data tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Hari libur dihapus"})
}
