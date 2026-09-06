#!/bin/bash
# Deploy script untuk DOM Cloud
# Cross-compile Go binary untuk Linux ARM64

set -e

echo "=== Pusaka Monitor Deploy Script ==="
echo ""

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "ERROR: Go tidak ditemukan. Install dari https://go.dev/dl/"
    exit 1
fi

echo "1. Building frontend (SvelteKit)..."
cd frontend
if [ -f "package.json" ]; then
    npm run build
fi
cd ..

echo "2. Cross-compiling Go binary untuk Linux ARM64..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -buildvcs=false -o pusaka-monitor .

echo "3. Binary berhasil dibuat: pusaka-monitor ($(du -h pusaka-monitor | cut -f1))"
echo ""
echo "=== Deploy ke DOM Cloud ==="
echo ""
echo "1. Upload folder ini ke DOM Cloud (zip atau git push)"
echo "2. Aktifkan fitur PostgreSQL di portal DOM Cloud"
echo "3. Set environment variables di portal:"
echo "   - PORT=8080"
echo "   - POSTGRES_HOST=localhost"
echo "   - POSTGRES_PORT=5432"
echo "   - POSTGRES_USER=postgres"
echo "   - POSTGRES_PASSWORD=<your-postgres-password>"
echo "   - POSTGRES_DB=pusaka_monitor"
echo "   - JWT_SECRET=<random-48-char-string>"
echo "   - SUPERADMIN_USERNAME=admin"
echo "   - SUPERADMIN_PASSWORD=<secure-password>"
echo "   - ENCRYPTION_KEY=<random-32-char-string>"
echo ""
echo "4. Set app_start_command: env PORT=\$PORT ./pusaka-monitor"
echo "5. Klik 'Terapkan' di portal"
echo ""
echo "=== Verifikasi ==="
echo "curl https://your-domain.sgp.dom.my.id/health"
echo ""
