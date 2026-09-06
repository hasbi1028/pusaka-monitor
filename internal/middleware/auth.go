package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
	"github.com/hasbiawal/pusaka-monitor/internal/session"
)

func AuthRequired(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public paths
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/auth/login") ||
			strings.HasPrefix(path, "/api/auth/register") ||
			strings.HasPrefix(path, "/login") ||
			strings.HasPrefix(path, "/register") ||
			strings.HasPrefix(path, "/pending") ||
			strings.HasPrefix(path, "/rejected") ||
			strings.HasPrefix(path, "/static") {
			c.Next()
			return
		}

		// Get token from cookie
		token, err := c.Cookie("session")
		if err != nil || token == "" {
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Unauthorized"})
				c.Abort()
				return
			}
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Verify JWT
		claims, err := session.VerifyToken(token, jwtSecret)
		if err != nil {
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Token tidak valid"})
				c.Abort()
				return
			}
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Check session in DB
		var dbSession models.Session
		if err := db.Where("token = ? AND expires_at > ?", token, gorm.Expr("NOW()")).First(&dbSession).Error; err != nil {
			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, models.ApiResponse{Error: "Session expired"})
				c.Abort()
				return
			}
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Superadmin bypass
		if claims.Role == "superadmin" {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("instansi_id", claims.InstansiID)
			c.Next()
			return
		}

		// Check instansi status
		if claims.InstansiID != "" {
			var instansi models.Instansi
			if err := db.First(&instansi, "id = ?", claims.InstansiID).Error; err == nil {
				if instansi.Status == "pending" {
					if strings.HasPrefix(path, "/api/") {
						c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Akun belum disetujui"})
						c.Abort()
						return
					}
					c.Redirect(http.StatusFound, "/pending")
					c.Abort()
					return
				}
				if instansi.Status == "rejected" {
					if strings.HasPrefix(path, "/api/") {
						c.JSON(http.StatusForbidden, models.ApiResponse{Error: "Akun ditolak"})
						c.Abort()
						return
					}
					c.Redirect(http.StatusFound, "/rejected")
					c.Abort()
					return
				}
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("instansi_id", claims.InstansiID)
		c.Next()
	}
}
