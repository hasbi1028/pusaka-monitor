package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type ScheduleHandler struct {
	DB *gorm.DB
}

// GET /api/schedules
func (h *ScheduleHandler) List(c *gin.Context) {
	var schedules []models.Schedule
	h.DB.Order("jam ASC, menit ASC").Find(&schedules)
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: schedules})
}

// POST /api/schedules
func (h *ScheduleHandler) Create(c *gin.Context) {
	var req struct {
		Jam   int    `json:"jam" binding:"required,min=0,max=23"`
		Menit int    `json:"menit" binding:"min=0,max=59"`
		Label string `json:"label" binding:"required"`
		Mode  string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Jam (0-23), menit (0-59), dan label wajib diisi"})
		return
	}
	if req.Mode == "" {
		req.Mode = "all"
	}
	schedule := models.Schedule{
		ID:    uuid.New().String(),
		Jam:   req.Jam,
		Menit: req.Menit,
		Label: req.Label,
		Aktif: true,
		Mode:  req.Mode,
	}
	if err := h.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal membuat jadwal"})
		return
	}
	c.JSON(http.StatusCreated, models.ApiResponse{Success: true, Data: schedule})
}

// PUT /api/schedules/:id
func (h *ScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var schedule models.Schedule
	if err := h.DB.First(&schedule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Jadwal tidak ditemukan"})
		return
	}
	var req struct {
		Jam   *int    `json:"jam"`
		Menit *int    `json:"menit"`
		Label *string `json:"label"`
		Aktif *bool   `json:"aktif"`
		Mode  *string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Format tidak valid"})
		return
	}
	updates := map[string]interface{}{}
	if req.Jam != nil {
		updates["jam"] = *req.Jam
	}
	if req.Menit != nil {
		updates["menit"] = *req.Menit
	}
	if req.Label != nil {
		updates["label"] = *req.Label
	}
	if req.Aktif != nil {
		updates["aktif"] = *req.Aktif
	}
	if req.Mode != nil {
		updates["mode"] = *req.Mode
	}
	updates["updated_at"] = time.Now()
	h.DB.Model(&schedule).Updates(updates)
	h.DB.First(&schedule, "id = ?", id)
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: schedule})
}

// DELETE /api/schedules/:id
func (h *ScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	result := h.DB.Delete(&models.Schedule{}, "id = ?", id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Jadwal tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Jadwal dihapus"})
}

// POST /api/schedules/:id/toggle
func (h *ScheduleHandler) Toggle(c *gin.Context) {
	id := c.Param("id")
	var schedule models.Schedule
	if err := h.DB.First(&schedule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "Jadwal tidak ditemukan"})
		return
	}
	newVal := !schedule.Aktif
	h.DB.Model(&schedule).Updates(map[string]interface{}{"aktif": newVal, "updated_at": time.Now()})
	schedule.Aktif = newVal
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: schedule})
}
