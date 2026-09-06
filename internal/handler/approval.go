package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

type ApprovalHandler struct {
	DB *gorm.DB
}

type ApproveRejectRequest struct {
	Action string `json:"action" binding:"required"`
	Reason string `json:"reason"`
}

// GET /superadmin/approval
func (h *ApprovalHandler) ApprovalPage(c *gin.Context) {
	var pending []models.Instansi
	h.DB.Where("status = ?", "pending").Order("created_at DESC").Find(&pending)

	var recent []models.Instansi
	h.DB.Where("status != ?", "pending").Order("created_at DESC").Limit(20).Find(&recent)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"pending": pending,
			"recent":  recent,
		},
	})
}

// POST /api/approval/:id
func (h *ApprovalHandler) ApproveReject(c *gin.Context) {
	id := c.Param("id")

	var req ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Action wajib diisi"})
		return
	}

	switch req.Action {
	case "approve":
		h.DB.Model(&models.Instansi{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     "approved",
			"approved_at": gorm.Expr("NOW()"),
		})
	case "reject":
		h.DB.Model(&models.Instansi{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":        "rejected",
			"rejected_at":   gorm.Expr("NOW()"),
			"reject_reason": req.Reason,
		})
	default:
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Action harus approve atau reject"})
		return
	}

	// Log
	log := models.ApprovalLog{
		InstansiID: id,
		Action:     req.Action,
		Actor:      "superadmin",
		Note:       req.Reason,
	}
	h.DB.Create(&log)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true})
}
