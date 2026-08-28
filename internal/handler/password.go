package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// Self-reset: user ganti password sendiri (butuh password lama)
// POST /api/me/reset-password
// body: { "old_password": "...", "new_password": "..." }
func (h *AuthHandler) ResetOwnPassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Password lama & baru wajib (min 4 karakter)"})
		return
	}

	var user models.User
	if h.DB.First(&user, "id = ?", userID).Error != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "User tidak ditemukan"})
		return
	}

	// Verifikasi password lama
	if !database.CheckPassword(req.OldPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Password lama salah"})
		return
	}

	hash, err := database.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal hash password"})
		return
	}
	h.DB.Model(&user).Update("password_hash", hash)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Password berhasil diubah"})
}

// Admin instansi reset password user di tenant-nya
// POST /api/admin/users/:id/reset-password
// body: { "new_password": "..." }
func (h *AuthHandler) AdminResetPassword(c *gin.Context) {
	adminInstansi, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	targetID := c.Param("id")

	// Cari target user
	var target models.User
	if h.DB.First(&target, "id = ?", targetID).Error != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "User tidak ditemukan"})
		return
	}

	// Admin instansi hanya boleh reset user di instansinya (kecuali superadmin)
	if role != "superadmin" && target.InstansiID != adminInstansi {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Tidak berwenang mereset user instansi lain"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Password baru wajib (min 4 karakter)"})
		return
	}

	hash, err := database.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal hash password"})
		return
	}
	h.DB.Model(&target).Update("password_hash", hash)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Password user berhasil direset"})
}

// GET /api/superadmin/users — list semua user (superadmin) + nama instansi
func (h *AuthHandler) ListUsers(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "superadmin" {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Hanya superadmin"})
		return
	}

	type UserRow struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		Role         string `json:"role"`
		InstansiNama string `json:"instansi_nama"`
		Aktif        bool   `json:"aktif"`
	}
	var rows []UserRow
	h.DB.Table("users u").
		Select("u.id, u.username, u.role, i.nama as instansi_nama, u.aktif").
		Joins("LEFT JOIN instansis i ON i.id = u.instansi_id").
		Order("u.role, u.username").
		Scan(&rows)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    rows,
	})
}

// Superadmin reset password user mana pun (tanpa password lama)
// POST /api/superadmin/users/:id/reset-password
// body: { "new_password": "..." }
func (h *AuthHandler) SuperadminResetPassword(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "superadmin" {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Hanya superadmin"})
		return
	}

	targetID := c.Param("id")
	var target models.User
	if h.DB.First(&target, "id = ?", targetID).Error != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "User tidak ditemukan"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Password baru wajib (min 4 karakter)"})
		return
	}

	hash, err := database.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal hash password"})
		return
	}
	h.DB.Model(&target).Update("password_hash", hash)

	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Message: "Password berhasil direset (superadmin)"})
}
