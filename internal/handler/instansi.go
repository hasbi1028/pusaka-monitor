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

// normalizeGroup — terima "1203...@g.us" atau "628xx" (atau kosong untuk
// hapus/fallback env). Kembalikan "" bila format tak dikenal.
func normalizeGroup(s string) string {
	t := ""
	for _, c := range s {
		if c == ' ' || c == '	' || c == '\n' || c == '-' {
			continue
		}
		t += string(c)
	}
	if t == "" {
		return ""
	}
	if len(t) > 5 && t[len(t)-5:] == "@g.us" {
		return t
	}
	digits := true
	for _, c := range t {
		if c < '0' || c > '9' {
			digits = false
			break
		}
	}
	if digits && (len(t) >= 9) {
		return t
	}
	return ""
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
			"wa_group":       instansi.WaGroup,
			"recap_aktif":    instansi.RecapAktif,
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
		WaGroup      *string `json:"wa_group"`
		RecapAktif   *bool   `json:"recap_aktif"`
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
	if req.WaGroup != nil {
		g := normalizeGroup(*req.WaGroup)
		if *req.WaGroup != "" && g == "" {
			c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "format grup WA tidak valid (pakai 628xx atau ...@g.us)"})
			return
		}
		updates["wa_group"] = g
	}
	if req.RecapAktif != nil {
		updates["recap_aktif"] = *req.RecapAktif
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
			"wa_group":       instansi.WaGroup,
			"recap_aktif":    instansi.RecapAktif,
		},
	})
}

// GET /api/superadmin/instansi — daftar semua instansi (khusus superadmin)
func (h *InstansiHandler) ListAll(c *gin.Context) {
	if role, _ := c.Get("role"); role != "superadmin" {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "khusus superadmin"})
		return
	}
	var list []models.Instansi
	h.DB.Order("nama").Find(&list)
	out := make([]gin.H, 0, len(list))
	for _, ins := range list {
		out = append(out, gin.H{
			"id": ins.ID, "nama": ins.Nama, "aktif": ins.Aktif,
			"status": ins.Status, "wa_group": ins.WaGroup, "recap_aktif": ins.RecapAktif,
		})
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: out})
}

// PUT /api/superadmin/instansi/:id — update rekap/grup per instansi (khusus superadmin)
func (h *InstansiHandler) UpdateOne(c *gin.Context) {
	if role, _ := c.Get("role"); role != "superadmin" {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "khusus superadmin"})
		return
	}
	var req struct {
		RecapAktif *bool   `json:"recap_aktif"`
		WaGroup    *string `json:"wa_group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tidak valid"})
		return
	}
	updates := map[string]interface{}{}
	if req.RecapAktif != nil {
		updates["recap_aktif"] = *req.RecapAktif
	}
	if req.WaGroup != nil {
		g := normalizeGroup(*req.WaGroup)
		if *req.WaGroup != "" && g == "" {
			c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "format grup WA tidak valid (pakai 628xx atau ...@g.us)"})
			return
		}
		updates["wa_group"] = g
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "tidak ada perubahan"})
		return
	}
	if err := h.DB.Model(&models.Instansi{}).Where("id = ?", c.Param("id")).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal menyimpan"})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Instansi diperbarui"})
}
