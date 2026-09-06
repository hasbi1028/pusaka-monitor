package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
	"github.com/hasbiawal/pusaka-monitor/internal/session"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	NamaInstansi string `json:"nama_instansi" binding:"required"`
	JnsInstansi  string `json:"jns_instansi" binding:"required"`
	Kabupaten    string `json:"kabupaten" binding:"required"`
	Provinsi     string `json:"provinsi" binding:"required"`
	Telepon      string `json:"telepon"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// isHTTPS deteksi apakah request via HTTPS
func isHTTPS(c *gin.Context) bool {
	return c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Username dan password wajib diisi"})
		return
	}

	// Find user
	var user models.User
	if err := h.DB.Where("username = ? AND aktif = ?", req.Username, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Username atau password salah"})
		return
	}

	// Check password
	if !database.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Username atau password salah"})
		return
	}

	// Check instansi status (if not superadmin)
	if user.Role != "superadmin" && user.InstansiID != "" {
		var instansi models.Instansi
		if err := h.DB.First(&instansi, "id = ?", user.InstansiID).Error; err == nil {
			if instansi.Status == "pending" {
				c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Akun belum disetujui. Menunggu persetujuan admin."})
				return
			}
			if instansi.Status == "rejected" {
				c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Pendaftaran ditolak. Hubungi admin."})
				return
			}
		}
	}

	// Create token
	token, err := session.CreateToken(user.ID, user.Username, user.InstansiID, user.Role, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal membuat token"})
		return
	}

	// Save session
	dbSession := models.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}
	h.DB.Create(&dbSession)

	// Update last login
	now := time.Now()
	h.DB.Model(&user).Update("last_login", &now)

	// Set cookie (secure jika HTTPS)
	secure := isHTTPS(c)
	c.SetCookie("session", token, 12*3600, "/", "", secure, true)

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"username":    user.Username,
			"role":        user.Role,
			"instansi_id": user.InstansiID,
		},
	})
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "Field wajib belum diisi"})
		return
	}

	// Create instansi (status: pending)
	instansiID := uuid.New().String()
	instansi := models.Instansi{
		ID:          instansiID,
		Nama:        req.NamaInstansi,
		JnsInstansi: req.JnsInstansi,
		Kabupaten:   req.Kabupaten,
		Provinsi:    req.Provinsi,
		Telepon:     req.Telepon,
		Status:      "pending",
		Aktif:       true,
	}
	if err := h.DB.Create(&instansi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal membuat instansi"})
		return
	}

	// Create user (role: admin)
	hash, err := database.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal hash password"})
		return
	}

	user := models.User{
		ID:           uuid.New().String(),
		InstansiID:   instansiID,
		Username:     req.Username,
		PasswordHash: hash,
		Role:         "admin",
		Aktif:        true,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal membuat user"})
		return
	}

	// Log approval
	logEntry := models.ApprovalLog{
		ID:         uuid.New().String(),
		InstansiID: instansiID,
		Action:     "registered",
		Actor:      "system",
	}
	h.DB.Create(&logEntry)

	c.JSON(http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "Pendaftaran dikirim. Menunggu persetujuan admin.",
		Data:    gin.H{"instansi_id": instansiID},
	})
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("session")
	if token != "" {
		h.DB.Where("token = ?", token).Delete(&models.Session{})
	}
	secure := isHTTPS(c)
	c.SetCookie("session", "", -1, "/", "", secure, true)
	c.JSON(http.StatusOK, models.ApiResponse{Success: true})
}

// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	instansiID, _ := c.Get("instansi_id")

	// Prevent stale auth data — browser must revalidate
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: gin.H{
			"user_id":     userID,
			"username":    username,
			"role":        role,
			"instansi_id": instansiID,
		},
	})
}
