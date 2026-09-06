package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/crypto"
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

	q := h.DB.Where("aktif = ?", true)
	if role != "superadmin" && instansiID != "" {
		q = h.DB.Where("instansi_id = ? AND aktif = ?", instansiID, true)
	}

	var pegawai []models.Pegawai
	q.Order("nama").Find(&pegawai)

	// Decrypt password_pusaka untuk response (hanya untuk admin sendiri)
	for i := range pegawai {
		if pegawai[i].PasswordPusaka != "" {
			// Sembunyikan password di response list
			pegawai[i].PasswordPusaka = "••••••••"
		}
	}

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

	// Encrypt password Pusaka sebelum simpan
	pwd := req.PasswordPusaka
	if pwd != "" {
		if encrypted, err := crypto.Encrypt(pwd); err == nil {
			pwd = encrypted
		}
	}

	pegawai := models.Pegawai{
		ID:             uuid.New().String(),
		InstansiID:     instStr,
		NIP:            req.NIP,
		Nama:           req.Nama,
		Jabatan:        req.Jabatan,
		Golongan:       req.Golongan,
		PasswordPusaka: pwd,
		Aktif:          true,
	}

	if err := h.DB.Create(&pegawai).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal menyimpan pegawai"})
		return
	}

	// Response: sembunyikan password
	pegawai.PasswordPusaka = "••••••••"
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
	if req.PasswordPusaka != "" && req.PasswordPusaka != "••••••••" {
		// Encrypt password baru
		if encrypted, err := crypto.Encrypt(req.PasswordPusaka); err == nil {
			existing.PasswordPusaka = encrypted
		} else {
			existing.PasswordPusaka = req.PasswordPusaka
		}
	}
	existing.UpdatedAt = time.Now()

	h.DB.Save(&existing)

	// Response: sembunyikan password
	existing.PasswordPusaka = "••••••••"
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: existing})
}

// DELETE /api/pegawai/:id
func (h *PegawaiHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	h.DB.Model(&models.Pegawai{}).Where("id = ?", id).Update("aktif", false)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Pegawai dinonaktifkan"})
}
