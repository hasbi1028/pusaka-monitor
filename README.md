# Pusaka Monitor

Monitor kehadiran Pusaka Kemenag — web app untuk memantau absensi pegawai secara otomatis.

## Fitur

- 🔐 **Multi-tenant** — Tiap instansi punya data terpisah
- 👥 **Role-based** — Superadmin, Admin Instansi
- 📊 **Dashboard** — Real-time rekap kehadiran
- 🔄 **Auto-scrape** — Jadwal otomatis ambil data dari Pusaka
- 📱 **Mobile-first** — UI responsif untuk Android
- 🔒 **AES-256-GCM** — Enkripsi password Pusaka
- 🗄️ **PostgreSQL** — Database production-ready

## Tech Stack

- **Backend**: Go + Gin + GORM
- **Database**: PostgreSQL (production) / SQLite (development)
- **Frontend**: SvelteKit 5 (adapter-static SPA)
- **Auth**: JWT + session cookies
- **Scraper**: HTTP client ke Pusaka v3 API

## Deploy ke DOM Cloud

### Prerequisites

1. Akun DOM Cloud (hasbi1028)
2. PostgreSQL diaktifkan di portal
3. Go 1.26+ terinstall

### Steps

```bash
# 1. Cross-compile untuk Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -buildvcs=false -o pusaka-monitor .

# 2. Upload ke DOM Cloud (zip atau git push)

# 3. Set env vars di portal DOM Cloud:
#    PORT=8080
#    POSTGRES_HOST=localhost
#    POSTGRES_PORT=5432
#    POSTGRES_USER=postgres
#    POSTGRES_PASSWORD=<your-password>
#    POSTGRES_DB=pusaka_monitor
#    JWT_SECRET=<random-48-char>
#    SUPERADMIN_USERNAME=admin
#    SUPERADMIN_PASSWORD=<secure>
#    ENCRYPTION_KEY=<random-32-char>

# 4. Set app_start_command: env PORT=$PORT ./pusaka-monitor

# 5. Klik "Terapkan"
```

### Verifikasi

```bash
# Health check
curl https://your-domain.sgp.dom.my.id/health

# Response: {"status":"ok","database":"postgresql"}
```

## Local Development

```bash
# 1. Copy .env.example ke .env
cp .env.example .env

# 2. Edit .env (isi PostgreSQL credentials atau SQLite path)

# 3. Build & run
go build -o pusaka-monitor.exe .
./pusaka-monitor.exe

# 4. Buka http://localhost:8080
```

## API Endpoints

### Public
- `POST /api/auth/login` — Login
- `POST /api/auth/register` — Register instansi baru
- `GET /health` — Health check

### Protected (butuh session cookie)
- `GET /api/dashboard` — Dashboard hari ini
- `GET /api/dashboard/bulan` — Rekap bulanan
- `GET /api/pegawai` — List pegawai
- `POST /api/scrape` — Trigger scrape
- `GET /api/schedules` — List jadwal

### Superadmin
- `GET /superadmin/approval` — List pending approval
- `POST /api/approval/:id` — Approve/reject instansi
- `POST /api/admin/concurrency` — Set worker concurrency

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | Server port |
| `DATABASE_URL` | No | - | PostgreSQL URL (overrides individual vars) |
| `POSTGRES_HOST` | Yes* | localhost | PostgreSQL host |
| `POSTGRES_PORT` | No | 5432 | PostgreSQL port |
| `POSTGRES_USER` | Yes* | postgres | PostgreSQL user |
| `POSTGRES_PASSWORD` | Yes* | - | PostgreSQL password |
| `POSTGRES_DB` | No | pusaka_monitor | Database name |
| `POSTGRES_SSLMODE` | No | disable | SSL mode |
| `JWT_SECRET` | Yes | - | JWT signing secret (min 32 chars) |
| `SUPERADMIN_USERNAME` | No | admin | Superadmin username |
| `SUPERADMIN_PASSWORD` | Yes | - | Superadmin password |
| `ENCRYPTION_KEY` | No | JWT_SECRET | AES encryption key (min 32 chars) |

*Required when `DATABASE_URL` is not set

## License

MIT
