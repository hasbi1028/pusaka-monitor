package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hasbiawal/pusaka-monitor/internal/config"
	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/handler"
	"github.com/hasbiawal/pusaka-monitor/internal/middleware"
	"github.com/hasbiawal/pusaka-monitor/templates"
)

func main() {
	// Load config
	cfg := config.Load()

	// Init database
	os.MkdirAll("data", 0755)
	db := database.Init(cfg.DBPath)

	// Seed superadmin
	database.SeedSuperAdmin(db, cfg.AdminUser, cfg.AdminPass)

	// Start background worker
	go handler.StartWorker(db)

	// Start scheduler: auto-scrape di jam 07:01 (pagi) & 17:00 (sore) WITA
	// + recover job stale tiap 60 detik. Bukan polling tiap 30 mnt
	// (hindari rate-limit/ban backend Pusaka).
	go handler.StartScheduler(db)

	// Handlers
	authHandler := &handler.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	approvalHandler := &handler.ApprovalHandler{DB: db}
	dashboardHandler := &handler.DashboardHandler{DB: db}
	pegawaiHandler := &handler.PegawaiHandler{DB: db}
	scrapeHandler := &handler.ScrapeHandler{DB: db}

	// Router
	r := gin.Default()

	// Static files
	r.Static("/static", "./static")

	// Public pages (templ)
	r.GET("/login", func(c *gin.Context) {
		templates.LoginPage("").Render(c.Request.Context(), c.Writer)
	})
	r.GET("/register", func(c *gin.Context) {
		templates.RegisterPage().Render(c.Request.Context(), c.Writer)
	})
	r.GET("/pending", func(c *gin.Context) {
		templates.PendingPage().Render(c.Request.Context(), c.Writer)
	})
	r.GET("/rejected", func(c *gin.Context) {
		templates.RejectedPage().Render(c.Request.Context(), c.Writer)
	})

	// Public API
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/register", authHandler.Register)

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	{
		// Auth API
		auth.POST("/api/auth/logout", authHandler.Logout)
		auth.GET("/api/auth/me", authHandler.Me)
		// Reset password
		auth.POST("/api/me/reset-password", authHandler.ResetOwnPassword)
		auth.POST("/api/admin/users/:id/reset-password", authHandler.AdminResetPassword)
		auth.POST("/api/superadmin/users/:id/reset-password", authHandler.SuperadminResetPassword)
		auth.GET("/api/superadmin/users", authHandler.ListUsers)

		// Dashboard API
		auth.GET("/api/dashboard", dashboardHandler.Dashboard)
		auth.GET("/api/dashboard/bulan", dashboardHandler.DashboardBulan)
		auth.GET("/api/dashboard/bulan-pegawai", dashboardHandler.DashboardBulanPegawai)
		auth.GET("/api/dashboard/bulan-detail", dashboardHandler.DashboardBulanDetail)

		// Pegawai API
		auth.GET("/api/pegawai", pegawaiHandler.List)
		auth.POST("/api/pegawai", pegawaiHandler.Create)
		auth.PUT("/api/pegawai/:id", pegawaiHandler.Update)
		auth.DELETE("/api/pegawai/:id", pegawaiHandler.Delete)

		// Scrape API
		auth.POST("/api/scrape", scrapeHandler.CreateJobs)
		auth.GET("/api/scrape/status", scrapeHandler.Status)
		auth.POST("/api/scrape/retry", scrapeHandler.RetryFailed)
		auth.POST("/api/scrape/cancel-all", scrapeHandler.CancelAll)
		auth.POST("/api/scrape/pegawai/:nip", scrapeHandler.ScrapeOne)
		auth.POST("/api/scrape/:id/cancel", scrapeHandler.CancelOne)

		// Superadmin API
		auth.GET("/superadmin/approval", approvalHandler.ApprovalPage)
		auth.POST("/api/approval/:id", approvalHandler.ApproveReject)
		// Import pegawai (superadmin) + Export CSV
		auth.POST("/api/admin/import-pegawai", pegawaiHandler.ImportPegawai)
		auth.GET("/api/admin/export-csv", dashboardHandler.ExportCSV)
		auth.GET("/api/admin/concurrency", scrapeHandler.GetConcurrencyHandler)
		auth.POST("/api/admin/concurrency", scrapeHandler.SetConcurrencyHandler)

		// Protected pages (templ)
		auth.GET("/", func(c *gin.Context) {
			username, _ := c.Get("username")
			templates.DashboardPage(username.(string)).Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/dashboard", func(c *gin.Context) {
			username, _ := c.Get("username")
			templates.DashboardPage(username.(string)).Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/pegawai", func(c *gin.Context) {
			templates.PegawaiPage().Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/scrape", func(c *gin.Context) {
			templates.ScrapePage().Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/settings", func(c *gin.Context) {
			username, _ := c.Get("username")
			role, _ := c.Get("role")
			templates.SettingsPage(username.(string), role.(string)).Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/profil", func(c *gin.Context) {
			templates.ProfilPage().Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/superadmin/users", func(c *gin.Context) {
			role, _ := c.Get("role")
			if role != "superadmin" {
				c.Redirect(http.StatusFound, "/")
				return
			}
			templates.UsersPage().Render(c.Request.Context(), c.Writer)
		})
		auth.GET("/approval", func(c *gin.Context) {
			templates.ApprovalPage().Render(c.Request.Context(), c.Writer)
		})
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Server running on http://localhost:%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
