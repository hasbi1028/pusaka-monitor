package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	_ "github.com/joho/godotenv/autoload"
	"github.com/gin-gonic/gin"
	"github.com/hasbiawal/pusaka-monitor/internal/backup"
	"github.com/hasbiawal/pusaka-monitor/internal/config"
	"github.com/hasbiawal/pusaka-monitor/internal/crypto"
	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/handler"
	"github.com/hasbiawal/pusaka-monitor/internal/middleware"
)

// AppVersion — tampil di /health & UI agar versi deploy bisa dibedakan.
const AppVersion = "v1.0.3"

// pidFile — kunci single-instance: start ganda (panel/manual) langsung ditolak.
const pidFile = "pusaka-monitor.pid"

func ensureSingleInstance() {
	data, err := os.ReadFile(pidFile)
	if err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
			// Cek apakah proses dengan PID itu benar-benar hidup DAN merupakan pusaka-monitor
			if exe, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil &&
				strings.Contains(string(exe), "pusaka-monitor") {
				// Pastikan bukan PID kita sendiri (restart tanpa cleanup)
				if pid != os.Getpid() {
					log.Fatalf("[FATAL] Proses pusaka-monitor lain (PID %d) masih jalan — "+
						"hentikan dulu agar jadwal tidak ganda.", pid)
				}
			}
			// PID mati atau stale → hapus file lama, lanjut start
			_ = os.Remove(pidFile)
		}
	}
	_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
}

func main() {
	ensureSingleInstance()
	// CLI flag: --create-superadmin (bootstrap mode)
	if len(os.Args) > 1 && os.Args[1] == "--create-superadmin" {
		cfg := config.Load()
		db := database.Init(cfg.DBURL)
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

	// Init database (PostgreSQL)
	db := database.Init(cfg.DBURL)

	// Seed superadmin
	database.SeedSuperAdmin(db, cfg.AdminUser, cfg.AdminPass)

	// Start auto-backup (tiap 24 jam)
	backup.StartAutoBackup(db, 24*time.Hour)

	// Start session cleanup (tiap 1 jam)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			n := database.CleanupExpiredSessions(db)
			if n > 0 {
				log.Printf("[SESSION] %d expired sessions dibersihkan", n)
			}
		}
	}()

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
	recapHandler := &handler.RecapHandler{DB: db}

	// Router
	r := gin.Default()

	// Health check (before NoRoute!)
	r.GET("/health", func(c *gin.Context) {
		// Check DB connection
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "DB connection failed"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "DB ping failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "postgresql", "version": AppVersion})
	})

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Serve SvelteKit static assets
	r.Static("/_app", "./frontend/build/_app")
	r.StaticFile("/favicon.svg", "./frontend/build/favicon.svg")
	// Legacy dist support (if any)
	r.Static("/assets", "./frontend/dist/assets")
	r.Static("/static", "./static")

	// Rate limiter untuk login & register: 30 request per menit per IP (longgar untuk E2E)
	authRL := middleware.RateLimit(30, time.Minute)

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
		auth.GET("/api/scrape/stream", scrapeHandler.StreamStatus)
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
		auth.GET("/api/superadmin/instansi", instansiHandler.ListAll)
		auth.PUT("/api/superadmin/instansi/:id", instansiHandler.UpdateOne)

		// Superadmin API
		auth.GET("/superadmin/approval", approvalHandler.ApprovalPage)
		auth.POST("/api/approval/:id", approvalHandler.ApproveReject)
		auth.POST("/api/admin/import-pegawai", pegawaiHandler.ImportPegawai)
		auth.GET("/api/admin/export-csv", dashboardHandler.ExportCSV)
		auth.GET("/api/admin/concurrency", scrapeHandler.GetConcurrencyHandler)
		auth.POST("/api/admin/concurrency", scrapeHandler.SetConcurrencyHandler)

		// Rekap harian (gambar WA)
		auth.GET("/api/admin/recap/preview", recapHandler.Preview)
		auth.POST("/api/admin/recap/send", recapHandler.SendNow)
		auth.POST("/api/recap/test", recapHandler.TestSend)
	}

	// SPA fallback: serve 200.html for all non-API routes (SvelteKit SPA)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Don't serve SPA for API routes
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/superadmin/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		// Serve SvelteKit SPA fallback
		if _, err := os.Stat("./frontend/build/200.html"); err == nil {
			c.File("./frontend/build/200.html")
			return
		}
		c.File("./frontend/dist/index.html")
	})

	log.Printf("Server running on http://localhost:%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
