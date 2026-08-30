package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type InstansiHandler struct {
	DB *gorm.DB
}

// GET /api/instansi/settings — ambil jam kerja instansi
func (h *InstansiHandler) GetSettings(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")

	var instansi models.Instansi
	if role == "superadmin" {
		// Superadmin lihat instansi pertama yang punya data
		h.DB.Where("status = 'approved'").First(&instansi)
	} else {
		h.DB.First(&instansi, "id = ?", instansiID)
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"jam_masuk":      instansi.JamMasuk,
			"jam_pulang":     instansi.JamPulang,
			"toleransi":      instansi.Toleransi,
			"jam_masuk_ram":  instansi.JamMasukRam,
			"jam_pulang_ram": instansi.JamPulangRam,
			"toleransi_ram":  instansi.ToleransiRam,
			"mode_ramadan":   instansi.ModeRamadan,
		},
	})
}

// PUT /api/instansi/settings — update jam kerja instansi
func (h *InstansiHandler) UpdateSettings(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")

	var instansi models.Instansi
	if role == "superadmin" {
		h.DB.Where("status = 'approved'").First(&instansi)
	} else {
		if err := h.DB.First(&instansi, "id = ?", instansiID).Error; err != nil {
			c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Instansi tidak ditemukan"})
			return
		}
	}

	var req struct {
		JamMasuk     *string `json:"jam_masuk"`
		JamPulang    *string `json:"jam_pulang"`
		Toleransi    *int    `json:"toleransi"`
		JamMasukRam  *string `json:"jam_masuk_ram"`
		JamPulangRam *string `json:"jam_pulang_ram"`
		ToleransiRam *int    `json:"toleransi_ram"`
		ModeRamadan  *bool   `json:"mode_ramadan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tidak valid"})
		return
	}

	updates := map[string]interface{}{}
	if req.JamMasuk != nil {
		updates["jam_masuk"] = *req.JamMasuk
	}
	if req.JamPulang != nil {
		updates["jam_pulang"] = *req.JamPulang
	}
	if req.Toleransi != nil {
		updates["toleransi"] = *req.Toleransi
	}
	if req.JamMasukRam != nil {
		updates["jam_masuk_ram"] = *req.JamMasukRam
	}
	if req.JamPulangRam != nil {
		updates["jam_pulang_ram"] = *req.JamPulangRam
	}
	if req.ToleransiRam != nil {
		updates["toleransi_ram"] = *req.ToleransiRam
	}
	if req.ModeRamadan != nil {
		updates["mode_ramadan"] = *req.ModeRamadan
	}

	h.DB.Model(&instansi).Updates(updates)
	h.DB.First(&instansi, "id = ?", instansi.ID)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Jam kerja diperbarui",
		Data: gin.H{
			"jam_masuk":      instansi.JamMasuk,
			"jam_pulang":     instansi.JamPulang,
			"toleransi":      instansi.Toleransi,
			"jam_masuk_ram":  instansi.JamMasukRam,
			"jam_pulang_ram": instansi.JamPulangRam,
			"toleransi_ram":  instansi.ToleransiRam,
			"mode_ramadan":   instansi.ModeRamadan,
		},
	})
}
