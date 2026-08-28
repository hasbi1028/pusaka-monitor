package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hasbiawal/pusaka-monitor/internal/database"
	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// PegawaiImport — struktur dari pegawai-backup-*.json
type PegawaiImport struct {
	NIP    string `json:"nip"`
	Nama   string `json:"nama"`
	Password string `json:"password"`
	Aktif  bool   `json:"aktif"`
}

type ImportFile struct {
	Format    string          `json:"format"`
	ExportedAt string         `json:"exportedAt"`
	Total     int             `json:"total"`
	Pegawai   []PegawaiImport `json:"pegawai"`
}

// POST /api/admin/import-pegawai
// Body: JSON file (multipart "file") ATAU raw JSON
// Buat instansi MTsN 2 Kolut jika belum ada, lalu import 35 pegawai.
func (h *PegawaiHandler) ImportPegawai(c *gin.Context) {
	var raw []byte
	var err error

	if file, errFile := c.FormFile("file"); errFile == nil {
		f, _ := file.Open()
		raw, err = io.ReadAll(f)
		f.Close()
	} else {
		raw, err = io.ReadAll(c.Request.Body)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Success: false, Error: "Gagal baca data"})
		return
	}

	var imp ImportFile
	if err := json.Unmarshal(raw, &imp); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Success: false, Error: "Format JSON tidak valid"})
		return
	}
	if len(imp.Pegawai) == 0 {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Success: false, Error: "Tidak ada pegawai di file"})
		return
	}

	// Cari/buat instansi MTsN 2 Kolut
	var instansi models.Instansi
	if err := h.DB.Where("nama = ?", "MTsN 2 Kolaka Utara").First(&instansi).Error; err != nil {
		instansi = models.Instansi{
			ID:          uuid.New().String(),
			Nama:        "MTsN 2 Kolaka Utara",
			JnsInstansi: "mtsn",
			Kabupaten:   "Kolaka Utara",
			Provinsi:    "Sulawesi Tenggara",
			Alamat:      "Jl. Lalume No. 42 Kel. Olo-Oloho",
			Telepon:     "",
			Email:       "mtsn.pakue@gmail.com",
			Aktif:       true,
			Status:      "approved",
			ApprovedAt:  &[]time.Time{time.Now()}[0],
		}
		h.DB.Create(&instansi)

		// Buat user admin untuk instansi ini
		hash, _ := database.HashPassword("admin123")
		h.DB.Create(&models.User{
			ID:           uuid.New().String(),
			InstansiID:   instansi.ID,
			Username:     "mtsn2kolut",
			PasswordHash: hash,
			Role:         "admin",
			Aktif:        true,
		})
	}

	// Import pegawai (upsert by NIP)
	imported := 0
	skipped := 0
	for _, p := range imp.Pegawai {
		var existing models.Pegawai
		if h.DB.Where("n_ip = ? AND instansi_id = ?", p.NIP, instansi.ID).First(&existing).Error == nil {
			skipped++
			continue
		}
		err := h.DB.Create(&models.Pegawai{
			ID:             uuid.New().String(),
			InstansiID:     instansi.ID,
			NIP:            p.NIP,
			Nama:           p.Nama,
			PasswordPusaka: p.Password,
			Aktif:          p.Aktif,
		}).Error
		if err == nil {
			imported++
		}
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Import selesai",
		Data: gin.H{
			"instansi_id": instansi.ID,
			"imported":    imported,
			"skipped":     skipped,
			"total_file":  len(imp.Pegawai),
		},
	})
}

// GET /api/admin/export-csv
// Export absensi ke CSV (filter bulan/tahun opsional)
func (h *DashboardHandler) ExportCSV(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	bulan := c.Query("bulan")
	tahun := c.Query("tahun")

	q := h.DB.Model(&models.Absensi{})
	if role != "superadmin" && instansiID != "" {
		q = q.Where("instansi_id = ?", instansiID)
	}
	if bulan != "" && tahun != "" {
		q = q.Where("tanggal LIKE ?", tahun+"-"+bulan+"-%")
	}

	var absens []models.Absensi
	q.Order("tanggal DESC, nama ASC").Find(&absens)

	// Build CSV (quote field yang mengandung koma/quote)
	csv := "NIP,Nama,Tanggal,JamMasuk,JamPulang,Status\n"
	quote := func(s string) string {
		if strings.ContainsAny(s, ",\"") {
			return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
		}
		return s
	}
	for _, a := range absens {
		csv += quote(a.NIP) + "," + quote(a.Nama) + "," + quote(a.Tanggal) + "," + quote(a.JamMasuk) + "," + quote(a.JamPulang) + "," + quote(a.Status) + "\n"
	}

	filename := "absensi"
	if bulan != "" && tahun != "" {
		filename += "_" + tahun + "-" + bulan
	}
	filename += ".csv"

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(http.StatusOK, csv)
}
