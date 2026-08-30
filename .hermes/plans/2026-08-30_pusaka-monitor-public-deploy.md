# Pusaka Monitor — Public Deployment Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Make pusaka-monitor ready for public free use by any Kemenag instansi (KUA, Madrasah, Kantor Kemenag) — not just MTsN 2 Kolut.

**Architecture:** Single-binary Go app + SvelteKit SPA + SQLite. Currently works internally; needs security hardening, generic onboarding, containerization, and documentation to be publicly deployable.

**Tech Stack:** Go 1.26, Gin, GORM, SQLite, SvelteKit 5, Tailwind v4, Docker

---

## Phase 1: Security Hardening (BLOCKER)

### Task 1: Wajibkan env vars kritis (JWT_SECRET, SUPERADMIN_PASSWORD)

**Objective:** App harus panic kalau JWT_SECRET atau SUPERADMIN_PASSWORD tidak diset — menghilangkan fallback hardcoded.

**Files:**
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/config/config.go`

**Step 1: Ganti config.go**

```go
package config

import (
	"log"
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("[FATAL] JWT_SECRET harus diset di .env atau environment. Contoh: JWT_SECRET=rahasia-banget-xyz123")
	}
	adminPass := os.Getenv("SUPERADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("[FATAL] SUPERADMIN_PASSWORD harus diset. Contoh: SUPERADMIN_PASSWORD=Sp3derman!")
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "data/presensi.db"),
		JWTSecret: jwtSecret,
		AdminUser: getEnv("SUPERADMIN_USERNAME", "admin"),
		AdminPass: adminPass,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

**Step 2: Build & verify**

```bash
cd ~/webapp/pusaka-monitor
go build -o pusaka-monitor.exe .
# Test: tanpa env var harus panic
JWT_SECRET="" SUPERADMIN_PASSWORD="" ./pusaka-monitor.exe
# Expected: FATAL exit dengan pesan jelas

# Test: dengan env var harus jalan
JWT_SECRET=test SUPERADMIN_PASSWORD=test123 ./pusaka-monitor.exe
# Expected: Server running on http://localhost:8080
```

**Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "fix(security): wajibkan JWT_SECRET & SUPERADMIN_PASSWORD env var"
```

---

### Task 2: Buat .env.example

**Objective:** User baru tahu env vars apa yang perlu diset.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/.env.example`

**Step 1: Buat file .env.example**

```
# ============================================
# Pusaka Monitor — Environment Variables
# ============================================
# Copy file ini ke .env lalu isi nilainya

# Port server (default: 8080)
PORT=8080

# Path SQLite database (default: data/presensi.db)
DB_PATH=data/presensi.db

# JWT signing secret — WAJIB, minimal 32 karakter acak
# Generate: python3 -c "import secrets; print(secrets.token_urlsafe(48))"
JWT_SECRET=

# Superadmin credentials — WAJIB
SUPERADMIN_USERNAME=admin
SUPERADMIN_PASSWORD=
```

**Step 2: Update .gitignore pastikan .env tidak ke-commit**

```bash
cat >> ~/webapp/pusaka-monitor/.gitignore << 'EOF'

# Environment secrets
.env
.env.local
.env.production
EOF
```

**Step 3: Commit**

```bash
git add .env.example .gitignore
git commit -m "chore: tambah .env.example untuk dokumentasi env vars"
```

---

### Task 3: Encrypt password_pusaka dengan AES-GCM

**Objective:** Password Pusaka pegawai tidak tersimpan plaintext di SQLite.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/crypto/aes.go`
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/handler/pegawai.go` (encrypt before save)
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/handler/scrape.go` (decrypt before scrape)
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/handler/import.go` (encrypt during import)
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/config/config.go` (add ENCRYPTION_KEY)

**Step 1: Buat crypto/aes.go**

```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var encryptionKey []byte

func Init(key string) error {
	if len(key) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY minimal 32 karakter, dapat: %d", len(key))
	}
	encryptionKey = []byte(key[:32])
	return nil
}

func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext terlalu pendek")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func IsEncrypted(s string) bool {
	// Ciphertext AES-GCM yang di-base64 selalu > 44 char (nonce 12 + minimal 16 tag)
	// Plaintext password biasanya 6-20 char
	if len(s) > 44 {
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err == nil && len(decoded) > 28 {
			return true
		}
	}
	return false
}
```

**Step 2: Tambah ENCRYPTION_KEY ke config.go**

```go
// Di struct Config:
EncryptionKey string

// Di Load():
encryptionKey := os.Getenv("ENCRYPTION_KEY")
if encryptionKey == "" {
	// Auto-generate dari JWT_SECRET (fallback untuk backward compat)
	encryptionKey = jwtSecret
	if len(encryptionKey) < 32 {
		log.Fatal("[FATAL] ENCRYPTION_KEY minimal 32 karakter")
	}
}
```

**Step 3: Encrypt di pegawai.go Create/Update**

```go
// Di PegawaiHandler.Create(), SEBELUM DB.Create(&pegawai):
if pegawai.PasswordPusaka != "" {
	encrypted, err := crypto.Encrypt(pegawai.PasswordPusaka)
	if err == nil {
		pegawai.PasswordPusaka = encrypted
	}
}
// Hal yang sama di Update()
```

**Step 4: Decrypt di scrape.go processJob()**

```go
// Di processJob(), SEBELUM scraper.NewClient():
pwd := pegawai.PasswordPusaka
if crypto.IsEncrypted(pwd) {
	decrypted, err := crypto.Decrypt(pwd)
	if err == nil {
		pwd = decrypted
	}
}
client := scraper.NewClient(pegawai.NIP, pegawai.Nama, pwd)
```

**Step 5: Encrypt di import.go**

```go
// Di import loop:
pwd := p.Password
if pwd != "" {
	encrypted, err := crypto.Encrypt(pwd)
	if err == nil {
		pwd = encrypted
	}
}
err := h.DB.Create(&models.Pegawai{
	PasswordPusaka: pwd,
	// ... other fields
}).Error
```

**Step 6: Tambah ENCRYPTION_KEY ke .env.example**

```
# Encryption key untuk password Pusaka (minimal 32 karakter)
# Generate: python3 -c "import secrets; print(secrets.token_urlsafe(32))"
# Jika kosong, fallback ke JWT_SECRET
ENCRYPTION_KEY=
```

**Step 7: Build & test**

```bash
go build -o pusaka-monitor.exe .
# Buat pegawai baru, cek DB:
sqlite3 data/presensi.db "SELECT password_pusaka FROM pegawais LIMIT 1"
# Expected: base64 string (bukan plaintext password)
```

**Step 8: Commit**

```bash
git add internal/crypto/ internal/handler/pegawai.go internal/handler/scrape.go internal/handler/import.go internal/config/config.go .env.example
git commit -m "feat(security): encrypt password_pusaka dengan AES-GCM"
```

---

### Task 4: Tambah rate limiting middleware

**Objective:** Brute force login dan abuse API dibatasi.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/middleware/ratelimit.go`
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/main.go`

**Step 1: Buat ratelimit.go**

```go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Cleanup setiap menit
	go func() {
		for range time.Tick(time.Minute) {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, times := range rl.requests {
		valid := times[:0]
		for _, t := range times {
			if now.Sub(t) <= rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	rl.requests[ip] = append(rl.requests[ip], now)
	count := 0
	for _, t := range rl.requests[ip] {
		if now.Sub(t) <= rl.window {
			count++
		}
	}
	return count <= rl.limit
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak request. Coba lagi dalam beberapa menit.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

**Step 2: Apply di main.go**

```go
// Login & Register: 5 request per menit per IP
authRL := middleware.RateLimit(5, time.Minute)

r.POST("/api/auth/login", authRL, authHandler.Login)
r.POST("/api/auth/register", authRL, authHandler.Register)
```

**Step 3: Build & test**

```bash
go build -o pusaka-monitor.exe .
# Test: curl 10x login cepat — harus kena 429 setelah 5 request
for i in {1..10}; do curl -s -o /dev/null -w "%{http_code} " -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"test","password":"test"}'; done
# Expected: 401 401 401 401 401 429 429 429 429 429
```

**Step 4: Commit**

```bash
git add internal/middleware/ratelimit.go main.go
git commit -m "feat(security): rate limiting 5 req/menit pada login/register"
```

---

### Task 5: Set cookie Secure flag & tambah SameSite

**Objective:** Cookie hanya dikirim via HTTPS, mencegah cookie theft.

**Files:**
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/handler/auth.go`

**Step 1: Update SetCookie calls**

```go
// Login handler - ganti baris:
c.SetCookie("session", token, 12*3600, "/", "", false, true)
// Menjadi:
secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
c.SetCookie("session", token, 12*3600, "/", "", secure, true)

// Logout handler - ganti baris:
c.SetCookie("session", "", -1, "/", "", false, true)
// Menjadi:
c.SetCookie("session", "", -1, "/", "", secure, true)
```

**Step 2: Commit**

```bash
git add internal/handler/auth.go
git commit -m "fix(security): cookie Secure flag berdasarkan HTTPS detection"
```

---

## Phase 2: Generic Onboarding (BLOCKER untuk publik)

### Task 6: Buat import generik (bukan hardcoded MTsN 2 Kolut)

**Objective:** Import pegawai harus bisa untuk instansi mana saja, bukan hanya MTsN 2 Kolut.

**Files:**
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/handler/import.go`

**Step 1: Ganti ImportPegawai handler**

```go
// Ganti fungsi ImportPegawai:

func (h *PegawaiHandler) ImportPegawai(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("instansi_id")
	instStr, _ := instansiID.(string)

	// Superadmin tanpa instansi → cari atau buat default
	if role == "superadmin" && (instStr == "" || instStr == "global") {
		// Cari instansi approved pertama
		var inst models.Instansi
		if h.DB.Where("status = 'approved'").First(&inst).Error == nil {
			instStr = inst.ID
		} else {
			c.JSON(http.StatusBadRequest, models.ApiResponse{
				Success: false,
				Error:   "Tidak ada instansi approved. Buat instansi dulu via Register.",
			})
			return
		}
	}

	// Baca file
	var raw []byte
	if file, errFile := c.FormFile("file"); errFile == nil {
		f, _ := file.Open()
		raw, _ = io.ReadAll(f)
		f.Close()
	} else {
		raw, _ = io.ReadAll(c.Request.Body)
	}

	var imp ImportFile
	if err := json.Unmarshal(raw, &imp); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Success: false, Error: "Format JSON tidak valid"})
		return
	}

	// Import ke instansi user (bukan hardcoded)
	imported, skipped := 0, 0
	for _, p := range imp.Pegawai {
		var existing models.Pegawai
		if h.DB.Where("n_ip = ? AND instansi_id = ?", p.NIP, instStr).First(&existing).Error == nil {
			skipped++
			continue
		}
		pwd := p.Password
		if pwd != "" {
			encrypted, _ := crypto.Encrypt(pwd)
			if encrypted != "" {
				pwd = encrypted
			}
		}
		if h.DB.Create(&models.Pegawai{
			ID:             uuid.New().String(),
			InstansiID:     instStr,
			NIP:            p.NIP,
			Nama:           p.Nama,
			PasswordPusaka: pwd,
			Aktif:          p.Aktif,
		}).Error == nil {
			imported++
		}
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Import selesai",
		Data:    gin.H{"instansi_id": instStr, "imported": imported, "skipped": skipped},
	})
}
```

**Step 2: Commit**

```bash
git add internal/handler/import.go
git commit -m "fix: import pegawai generik berdasarkan instansi_id user"
```

---

### Task 7: Tambah registrasi superadmin (opsional, via CLI flag)

**Objective:** Operator bisa create superadmin pertama tanpa hardcode.

**Files:**
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/main.go`

**Step 1: Tambah flag --create-superadmin**

```go
// Di main(), SEBELUM server start:

if len(os.Args) > 1 && os.Args[1] == "--create-superadmin" {
	username := getEnvOrDefault("SUPERADMIN_USERNAME", "admin")
	password := cfg.AdminPass
	database.SeedSuperAdmin(db, username, password)
	log.Printf("[ADMIN] Superadmin '%s' dibuat/diperiksa. Jalankan tanpa --create-superadmin untuk start server.", username)
	return
}
```

**Step 2: Commit**

```bash
git add main.go
git commit -m "feat: --create-superadmin CLI flag untuk bootstrap"
```

---

## Phase 3: Deployment Infrastructure

### Task 8: Buat Dockerfile multi-stage

**Objective:** User bisa deploy dengan 1 command: docker build && docker run.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/Dockerfile`
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/.dockerignore`

**Step 1: Buat Dockerfile**

```dockerfile
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o pusaka-monitor .

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Makassar
WORKDIR /app
COPY --from=backend /app/pusaka-monitor .
COPY --from=backend /app/frontend/dist ./frontend/dist
COPY --from=backend /app/static ./static
RUN mkdir -p data

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["./pusaka-monitor"]
```

**Step 2: Buat .dockerignore**

```
.git
node_modules
*.exe
*.exe~
logs/
data/
frontend/node_modules
frontend/.svelte-kit
```

**Step 3: Build & test**

```bash
cd ~/webapp/pusaka-monitor
docker build -t pusaka-monitor .
docker run -p 8080:8080 \
  -e JWT_SECRET=test123456789012345678901234567890 \
  -e SUPERADMIN_PASSWORD=Admin123! \
  -v pusaka-data:/app/data \
  pusaka-monitor
# Expected: Server running on http://localhost:8080
# Buka http://localhost:8080 — harus ada halaman login
```

**Step 4: Commit**

```bash
git add Dockerfile .dockerignore
git commit -m "feat(deploy): Dockerfile multi-stage untuk deployment publik"
```

---

### Task 9: Buat docker-compose.yml untuk 1-click deploy

**Objective:** Operator tinggal `docker compose up -d` dan langsung jalan.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/docker-compose.yml`

**Step 1: Buat docker-compose.yml**

```yaml
services:
  app:
    build: .
    ports:
      - "${PORT:-8080}:8080"
    env_file: .env
    volumes:
      - pusaka-data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3

volumes:
  pusaka-data:
```

**Step 2: Commit**

```bash
git add docker-compose.yml
git commit -m "feat(deploy): docker-compose untuk 1-click deployment"
```

---

### Task 10: Tambah auto-backup SQLite

**Objective:** Backup database otomatis tiap hari, simpan 7 hari terakhir.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/internal/backup/backup.go`
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/main.go`

**Step 1: Buat backup/backup.go**

```go
package backup

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func StartAutoBackup(dbPath string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			backup(dbPath)
		}
	}()
}

func backup(dbPath string) {
	dir := filepath.Join(filepath.Dir(dbPath), "backups")
	os.MkdirAll(dir, 0755)
	timestamp := time.Now().Format("2006-01-02_150405")
	dst := filepath.Join(dir, fmt.Sprintf("presensi_%s.db", timestamp))

	src, err := os.Open(dbPath)
	if err != nil {
		log.Printf("[BACKUP] Gagal buka DB: %v", err)
		return
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		log.Printf("[BACKUP] Gagal buat backup: %v", err)
		return
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, src); err != nil {
		log.Printf("[BACKUP] Gagal copy: %v", err)
		return
	}

	// Cleanup: hapus backup lebih dari 7 hari
	entries, _ := os.ReadDir(dir)
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	for _, e := range entries {
		info, err := e.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}

	log.Printf("[BACKUP] Database dibackup: %s", dst)
}
```

**Step 2: Tambah ke main.go**

```go
import "github.com/hasbiawal/pusaka-monitor/internal/backup"

// Setelah db init:
backup.StartAutoBackup(cfg.DBPath, 24*time.Hour)
```

**Step 3: Commit**

```bash
git add internal/backup/backup.go main.go
git commit -m "feat: auto-backup SQLite tiap 24 jam (simpan 7 hari)"
```

---

## Phase 4: Documentation (WAJIB untuk publik)

### Task 11: Buat README.md lengkap

**Objective:** User baru bisa setup, configure, dan deploy dalam 10 menit.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/README.md`

**Step 1: Buat README.md**

```markdown
# Pusaka Monitor

Monitoring presensi Pusaka Kemenag secara real-time. Aplikasi ini
HANYA membaca riwayat presensi dari Pusaka v3 — tidak mengubah
atau membuat data di sistem Kemenag.

## Fitur

- Dashboard harian & bulanan dengan statistik kehadiran
- Scrape real-time (semua / belum masuk / belum pulang)
- Auto-scrape terjadwal (DB-driven, configurable)
- Worker pool concurrent (default 8, max 100)
- Multi-tenant: setiap instansi punya data terpisah
- Approval workflow: superadmin setujui pendaftaran instansi
- Mode Ramadan: jam kerja otomatis berubah
- Export CSV
- Auto-backup database

## Tech Stack

- **Backend:** Go + Gin + GORM + SQLite
- **Frontend:** SvelteKit 5 + Tailwind CSS v4
- **Auth:** JWT + session DB (cookie-based)

## Quick Start (Docker)

```bash
# 1. Clone
git clone https://github.com/hasbiawal/pusaka-monitor.git
cd pusaka-monitor

# 2. Buat .env
cp .env.example .env
# Edit .env, isi JWT_SECRET dan SUPERADMIN_PASSWORD

# 3. Build & Run
docker compose up -d

# 4. Buka http://localhost:8080
# Login: admin / (password dari .env SUPERADMIN_PASSWORD)
```

## Quick Start (Manual)

```bash
# 1. Install dependencies
# Go 1.24+ dan Node.js 22+ harus terinstall

# 2. Build frontend
cd frontend && npm ci && npm run build && cd ..

# 3. Build backend
go build -o pusaka-monitor .

# 4. Buat .env
cp .env.example .env
# Edit .env

# 5. Run
./pusaka-monitor
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | No | 8080 | Port server |
| DB_PATH | No | data/presensi.db | Path SQLite |
| JWT_SECRET | **YES** | - | Secret untuk signing JWT |
| SUPERADMIN_PASSWORD | **YES** | - | Password superadmin |
| SUPERADMIN_USERNAME | No | admin | Username superadmin |
| ENCRYPTION_KEY | No | (JWT_SECRET) | Key untuk encrypt password Pusaka |

## Konfigurasi

### Jam Kerja
- **Normal:** Jam masuk/pulang standar + toleransi telat
- **Ramadan:** Jam masuk/pulang puasa + toleransi terpisah
- Toggle mode Ramadan di Settings

### Auto-Scrape
Buat jadwal di Settings > Jadwal Auto Scrape:
- **Mode All:** Scrape semua pegawai
- **Mode Belum Masuk:** Hanya yang belum ada data masuk
- **Mode Belum Pulang:** Hanya yang sudah masuk tapi belum pulang

### Worker Concurrency
Atur jumlah worker paralel di Settings > Worker Scrape.
Max 100. Default 8. Rate limit Pusaka: 60 req/jam/NIP.

## Deployment

### VPS (Linux)
```bash
# Build untuk Linux
GOOS=linux GOARCH=amd64 go build -o pusaka-monitor .

# Upload ke VPS
scp pusaka-monitor user@vps:/app/

# Setup systemd service
# (lihat wiki/DEPLOYMENT.md)
```

### Docker + Reverse Proxy (Nginx/Caddy)
```bash
docker compose up -d
# Akses via domain dengan SSL
```

## License

MIT
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: README.md lengkap untuk deployment publik"
```

---

### Task 12: Buat LICENSE file

**Objective:** App bisa di-publish ke GitHub secara legal.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/LICENSE`

**Step 1: Buat LICENSE (MIT)**

```
MIT License

Copyright (c) 2026 Hasbi Awal

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

**Step 2: Commit**

```bash
git add LICENSE
git commit -m "docs: tambah MIT License"
```

---

## Phase 5: Frontend Improvements

### Task 13: Tambah breadcrumb navigation

**Objective:** User tahu posisi mereka di navigasi.

**Files:**
- Create: `C:/Users/LENOVO/webapp/pusaka-monitor/frontend/src/lib/components/Breadcrumb.svelte`
- Modify: Setiap page file (dashboard, pegawai, scrape, settings, laporan, profil, approval, superadmin/users)

**Step 1: Buat Breadcrumb.svelte**

```svelte
<script>
  let { items = [] } = $props();
</script>

<nav class="flex items-center gap-1.5 text-xs text-gray-400 mb-4">
  {#each items as item, i}
    {#if i > 0}
      <span class="text-gray-300">/</span>
    {/if}
    {#if item.href}
      <a href={item.href} class="hover:text-blue-600 transition-colors">{item.label}</a>
    {:else}
      <span class="text-gray-700 font-medium">{item.label}</span>
    {/if}
  {/each}
</nav>
```

**Step 2: Tambah ke setiap page**

Contoh di dashboard:
```svelte
<script>
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
</script>

<Layout title="Dashboard" activePage="dashboard">
  <Breadcrumb items={[
    { label: 'Beranda' },
    { label: 'Dashboard' }
  ]} />
  <!-- konten existing -->
</Layout>
```

Contoh di pegawai:
```svelte
<Breadcrumb items={[
  { label: 'Beranda', href: '/dashboard' },
  { label: 'Pegawai' }
]} />
```

**Step 3: Build & test**

```bash
cd frontend && npm run build
# Buka setiap halaman — harus ada breadcrumb di atas konten
```

**Step 4: Commit**

```bash
git add frontend/src/lib/components/Breadcrumb.svelte frontend/src/routes/
git commit -m "feat(ui): tambah breadcrumb navigation"
```

---

### Task 14: Tambah loading skeleton di dashboard

**Objective:** User tidak melihat halaman kosong saat data dimuat.

**Files:**
- Modify: `C:/Users/LENOVO/webapp/pusaka-monitor/frontend/src/routes/dashboard/+page.svelte`

**Step 1: Tambah skeleton component**

```svelte
<!-- Ganti stat cards loading state dengan skeleton -->
{#if loading}
  <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-4">
    {#each Array(6) as _}
      <div class="bg-white rounded-xl border border-gray-200 p-3 animate-pulse">
        <div class="h-6 bg-gray-200 rounded w-12 mb-2"></div>
        <div class="h-3 bg-gray-100 rounded w-16"></div>
      </div>
    {/each}
  </div>
{/if}
```

**Step 2: Commit**

```bash
git add frontend/src/routes/dashboard/+page.svelte
git commit -m "feat(ui): loading skeleton di dashboard"
```

---

## Phase 6: Push to GitHub & Deploy

### Task 15: Push ke GitHub public repo

**Objective:** App tersedia untuk publik di GitHub.

**Files:**
- Modify: `.gitignore` (pastikan tidak ada .env, data/, *.db)

**Step 1: Update .gitignore**

```
# Environment
.env
.env.local
.env.production

# Data
data/
logs/
*.db
*.db-shm
*.db-wal

# Build artifacts
*.exe
*.exe~
pusaka-monitor

# Node
node_modules/
frontend/node_modules/
frontend/.svelte-kit/
frontend/build/
```

**Step 2: Create repo & push**

```bash
cd ~/webapp/pusaka-monitor
git remote add origin https://github.com/hasbiawal/pusaka-monitor.git
git push -u origin main
```

**Step 3: Create GitHub release v1.0.0**

```bash
git tag -a v1.0.0 -m "Public release: Pusaka Monitor v1.0.0"
git push origin v1.0.0
```

---

## Summary: File Changes

| Task | Files Changed |
|------|--------------|
| 1 | `internal/config/config.go` |
| 2 | `.env.example`, `.gitignore` |
| 3 | `internal/crypto/aes.go` (new), `handler/pegawai.go`, `handler/scrape.go`, `handler/import.go`, `config/config.go` |
| 4 | `internal/middleware/ratelimit.go` (new), `main.go` |
| 5 | `internal/handler/auth.go` |
| 6 | `internal/handler/import.go` |
| 7 | `main.go` |
| 8 | `Dockerfile` (new), `.dockerignore` (new) |
| 9 | `docker-compose.yml` (new) |
| 10 | `internal/backup/backup.go` (new), `main.go` |
| 11 | `README.md` (new) |
| 12 | `LICENSE` (new) |
| 13 | `frontend/src/lib/components/Breadcrumb.svelte` (new), all page files |
| 14 | `frontend/src/routes/dashboard/+page.svelte` |
| 15 | `.gitignore` |

## Execution Order

```
Phase 1 (Security):     Task 1 → 2 → 3 → 4 → 5
Phase 2 (Onboarding):   Task 6 → 7
Phase 3 (Deploy):       Task 8 → 9 → 10
Phase 4 (Docs):         Task 11 → 12
Phase 5 (Frontend):     Task 13 → 14
Phase 6 (Publish):      Task 15
```

## Estimasi Waktu

| Phase | Estimasi |
|-------|----------|
| Phase 1: Security | 2-3 jam |
| Phase 2: Onboarding | 1 jam |
| Phase 3: Deploy | 1-2 jam |
| Phase 4: Docs | 1 jam |
| Phase 5: Frontend | 1 jam |
| Phase 6: Publish | 30 menit |
| **Total** | **6-8 jam** |
