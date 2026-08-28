package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type PegawaiHandler struct {
	DB *gorm.DB
}

type PegawaiRequest struct {
	NIP            string `json:"nip" binding:"required"`
	Nama           string `json:"nama" binding:"required"`
	Jabatan        string `json:"jabatan"`
	Golongan       string `json:"golongan"`
	PasswordPusaka string `json:"password_pusaka"`
}

// GET /api/pegawai
func (h *PegawaiHandler) List(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")

	q := h.DB.Where("aktif = 1")
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("instansi_id = ? AND aktif = 1", instansiID)
	}

	var pegawai []models.Pegawai
	q.Order("nama").Find(&pegawai)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: pegawai})
}

// POST /api/pegawai
func (h *PegawaiHandler) Create(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")

	// Superadmin tanpa instansi → pakai "global"
	if role == "superadmin" && (instansiID == "" || instansiID == nil) {
		instansiID = "global"
	}
	instStr, _ := instansiID.(string)

	var req PegawaiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "NIP dan nama wajib diisi"})
		return
	}

	pegawai := models.Pegawai{
		ID:             uuid.New().String(),
		InstansiID:     instStr,
		NIP:            req.NIP,
		Nama:           req.Nama,
		Jabatan:        req.Jabatan,
		Golongan:       req.Golongan,
		PasswordPusaka: req.PasswordPusaka,
		Aktif:          true,
	}

	if err := h.DB.Create(&pegawai).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal menyimpan pegawai"})
		return
	}

	c.JSON(http.StatusCreated, models.ApiResponse{Success: true, Data: pegawai})
}

// PUT /api/pegawai/:id
func (h *PegawaiHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var existing models.Pegawai
	if err := h.DB.First(&existing, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Pegawai tidak ditemukan"})
		return
	}

	var req PegawaiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "NIP dan nama wajib diisi"})
		return
	}

	existing.NIP = req.NIP
	existing.Nama = req.Nama
	existing.Jabatan = req.Jabatan
	existing.Golongan = req.Golongan
	if req.PasswordPusaka != "" {
		existing.PasswordPusaka = req.PasswordPusaka
	}
	existing.UpdatedAt = time.Now()

	h.DB.Save(&existing)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: existing})
}

// DELETE /api/pegawai/:id
func (h *PegawaiHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	h.DB.Model(&models.Pegawai{}).Where("id = ?", id).Update("aktif", false)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Pegawai dinonaktifkan"})
}
