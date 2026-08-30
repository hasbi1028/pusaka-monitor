package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hasbiawal/pusaka-monitor/internal/backup"
	"github.com/hasbiawal/pusaka-monitor/internal/config"
	"github.com/hasbiawal/pusaka-monitor/internal/crypto"
	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/handler"
	"github.com/hasbiawal/pusaka-monitor/internal/middleware"
)

func main() {
	// CLI flag: --create-superadmin (bootstrap mode)
	if len(os.Args) > 1 && os.Args[1] == "--create-superadmin" {
		cfg := config.Load()
		database.Init(cfg.DBPath)
		db := database.Init(cfg.DBPath)
		database.SeedSuperAdmin(db, cfg.AdminUser, cfg.AdminPass)
		log.Printf("[ADMIN] Superadmin '%s' dibuat/diperiksa. Jalankan tanpa --create-superadmin untuk start server.", cfg.AdminUser)
		return
	}

	// Load config
	cfg := config.Load()

	// Init encryption
	if err := crypto.Init(cfg.EncryptionKey); err != nil {
		log.Fatalf("[FATAL] Gagal init encryption: %v", err)
	}

	// Init database
	os.MkdirAll("data", 0755)
	db := database.Init(cfg.DBPath)

	// Seed superadmin
	database.SeedSuperAdmin(db, cfg.AdminUser, cfg.AdminPass)

	// Start auto-backup (tiap 24 jam)
	backup.StartAutoBackup(cfg.DBPath, 24*time.Hour)

	// Start background worker
	go handler.StartWorker(db)

	// Start scheduler
	go handler.StartScheduler(db)

	// Handlers
	authHandler := &handler.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	approvalHandler := &handler.ApprovalHandler{DB: db}
	dashboardHandler := &handler.DashboardHandler{DB: db}
	pegawaiHandler := &handler.PegawaiHandler{DB: db}
	scrapeHandler := &handler.ScrapeHandler{DB: db}
	scheduleHandler := &handler.ScheduleHandler{DB: db}
	instansiHandler := &handler.InstansiHandler{DB: db}

	// Router
	r := gin.Default()

	// Serve static files from frontend/dist
	r.Static("/assets", "./frontend/dist/assets")
	r.Static("/static", "./static")

	// Rate limiter untuk login & register: 5 request per menit per IP
	authRL := middleware.RateLimit(5, time.Minute)

	// Public API
	r.POST("/api/auth/login", authRL, authHandler.Login)
	r.POST("/api/auth/register", authRL, authHandler.Register)

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	{
		// Auth API
		auth.POST("/api/auth/logout", authHandler.Logout)
		auth.GET("/api/auth/me", authHandler.Me)
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

		// Schedule API
		auth.GET("/api/schedules", scheduleHandler.List)
		auth.POST("/api/schedules", scheduleHandler.Create)
		auth.PUT("/api/schedules/:id", scheduleHandler.Update)
		auth.DELETE("/api/schedules/:id", scheduleHandler.Delete)
		auth.POST("/api/schedules/:id/toggle", scheduleHandler.Toggle)

		// Instansi Settings API
		auth.GET("/api/instansi/settings", instansiHandler.GetSettings)
		auth.PUT("/api/instansi/settings", instansiHandler.UpdateSettings)

		// Superadmin API
		auth.GET("/superadmin/approval", approvalHandler.ApprovalPage)
		auth.POST("/api/approval/:id", approvalHandler.ApproveReject)
		auth.POST("/api/admin/import-pegawai", pegawaiHandler.ImportPegawai)
		auth.GET("/api/admin/export-csv", dashboardHandler.ExportCSV)
		auth.GET("/api/admin/concurrency", scrapeHandler.GetConcurrencyHandler)
		auth.POST("/api/admin/concurrency", scrapeHandler.SetConcurrencyHandler)
	}

	// SPA fallback: serve index.html for all non-API, non-static routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Don't serve SPA for API routes
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/superadmin/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./frontend/dist/index.html")
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Server running on http://localhost:%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
