package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type CutiHandler struct {
	DB *gorm.DB
}

// GET /api/cuti?tanggal=YYYY-MM-DD — daftar cuti yang mencakup tanggal
// (default: hari ini WITA). Tanpa tanggal: semua cuti aktif/mendatang.
func (h *CutiHandler) List(c *gin.Context) {
	instansiID, role := liburScope(c)
	tanggal := c.DefaultQuery("tanggal", time.Now().UTC().Add(8*time.Hour).Format("2006-01-02"))

	q := h.DB.Order("tanggal_mulai ASC")
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ?", instansiID)
	}
	var list []models.Cuti
	// Tampilkan cuti yang mencakup tanggal (termasuk yang sedang berjalan)
	q.Where("tanggal_mulai <= ? AND tanggal_akhir >= ?", tanggal, tanggal).Find(&list)
	if list == nil {
		list = []models.Cuti{}
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: list})
}

// POST /api/cuti  body: {"nip":"...","tanggal_mulai":"...","tanggal_akhir":"...","keterangan":"..."}
func (h *CutiHandler) Create(c *gin.Context) {
	instansiID, role := liburScope(c)
	var req struct {
		NIP           string `json:"nip" binding:"required"`
		TanggalMulai  string `json:"tanggal_mulai" binding:"required"`
		TanggalAkhir  string `json:"tanggal_akhir" binding:"required"`
		Keterangan    string `json:"keterangan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "NIP, tanggal_mulai, tanggal_akhir wajib diisi"})
		return
	}
	if _, err := time.Parse("2006-01-02", req.TanggalMulai); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tanggal_mulai harus YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", req.TanggalAkhir); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tanggal_akhir harus YYYY-MM-DD"})
		return
	}
	if req.TanggalAkhir < req.TanggalMulai {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "tanggal_akhir tidak boleh sebelum tanggal_mulai"})
		return
	}

	// Pegawai harus ada; admin instansi hanya untuk instansinya
	var pegawai models.Pegawai
	pq := h.DB.Where("n_ip = ? AND aktif = ?", req.NIP, true)
	if role != "superadmin" && instansiID != "" {
		pq = pq.Where("instansi_id = ?", instansiID)
	}
	if err := pq.First(&pegawai).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Pegawai tidak ditemukan"})
		return
	}

	rec := models.Cuti{
		ID:           uuid.New().String(),
		InstansiID:   pegawai.InstansiID,
		NIP:          pegawai.NIP,
		Nama:         pegawai.Nama,
		TanggalMulai: req.TanggalMulai,
		TanggalAkhir: req.TanggalAkhir,
		Keterangan:   req.Keterangan,
		CreatedAt:    time.Now(),
	}
	if err := h.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal menyimpan cuti"})
		return
	}
	c.JSON(http.StatusCreated, models.ApiResponse{Success: true, Data: rec})
}

// DELETE /api/cuti/:id
func (h *CutiHandler) Delete(c *gin.Context) {
	instansiID, role := liburScope(c)
	id := c.Param("id")
	q := h.DB.Where("id = ?", id)
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ?", instansiID)
	}
	if q.Delete(&models.Cuti{}).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Data cuti tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Cuti dihapus"})
}
