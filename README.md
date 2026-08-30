# Pusaka Monitor

Monitoring presensi Pusaka Kemenag secara real-time. Aplikasi ini HANYA membaca riwayat presensi dari Pusaka v3 — tidak mengubah atau membuat data di sistem Kemenag.

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

- **Backend:** Go + Gin + GORM + SQLite (WAL mode)
- **Frontend:** SvelteKit 5 + Tailwind CSS v4
- **Auth:** JWT + session DB (cookie-based)
- **Security:** AES-256-GCM encryption untuk password Pusaka

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
# Prerequisites: Go 1.24+ dan Node.js 22+

# 1. Build frontend
cd frontend && npm ci && npm run build && cd ..

# 2. Build backend
go build -o pusaka-monitor .

# 3. Buat .env
cp .env.example .env
# Edit isi JWT_SECRET dan SUPERADMIN_PASSWORD

# 4. Run
./pusaka-monitor
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Port server |
| `DB_PATH` | No | `data/presensi.db` | Path SQLite database |
| `JWT_SECRET` | **YES** | - | Secret untuk signing JWT (min 32 karakter) |
| `SUPERADMIN_PASSWORD` | **YES** | - | Password superadmin |
| `SUPERADMIN_USERNAME` | No | `admin` | Username superadmin |
| `ENCRYPTION_KEY` | No | (JWT_SECRET) | Key untuk encrypt password Pusaka (min 32 karakter) |

### Generate secrets

```bash
# JWT_SECRET
python3 -c "import secrets; print(secrets.token_urlsafe(48))"

# ENCRYPTION_KEY
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

## Bootstrap Superadmin

```bash
# Buat superadmin tanpa start server
./pusaka-monitor --create-superadmin
```

## Konfigurasi

### Jam Kerja

- **Normal:** Jam masuk/pulang standar + toleransi telat (default 15 menit)
- **Ramadan:** Jam masuk/pulang puasa + toleransi terpisah
- Toggle mode Ramadan di Settings

### Auto-Scrape

Buat jadwal di Settings > Jadwal Auto Scrape:

| Mode | Deskripsi |
|------|-----------|
| `all` | Scrape semua pegawai |
| `belum_masuk` | Hanya yang belum ada data masuk |
| `belum_pulang` | Hanya yang sudah masuk tapi belum pulang |

### Worker Concurrency

Atur jumlah worker paralel di Settings > Worker Scrape.

- Default: 8 worker
- Max: 100 worker
- Rate limit Pusaka: 60 req/jam/NIP

### Import Pegawai

Format JSON untuk import:

```json
{
  "format": "pusaka-monitor",
  "total": 2,
  "pegawai": [
    {"nip": "199210282025211017", "nama": "Hasbi Awal", "password": "xxx", "aktif": true},
    {"nip": "123456789012345678", "nama": "Budi Santoso", "password": "yyy", "aktif": true}
  ]
}
```

## Deployment

### VPS (Linux)

```bash
# Build untuk Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pusaka-monitor .

# Upload ke VPS
scp pusaka-monitor user@vps:/app/

# Setup systemd service
sudo tee /etc/systemd/system/pusaka-monitor.service << 'EOF'
[Unit]
Description=Pusaka Monitor
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/app
ExecStart=/app/pusaka-monitor
Restart=always
EnvironmentFile=/app/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable pusaka-monitor
sudo systemctl start pusaka-monitor
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name monitor.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name monitor.example.com;

    ssl_certificate /etc/letsencrypt/live/monitor.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/monitor.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## License

MIT
