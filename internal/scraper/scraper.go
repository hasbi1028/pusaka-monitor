package scraper

import (
	"fmt"
	"strings"
	"time"
)

var bulanMap = map[string]string{
	"Januari": "01", "Februari": "02", "Maret": "03", "April": "04",
	"Mei": "05", "Juni": "06", "Juli": "07", "Agustus": "08",
	"September": "09", "Oktober": "10", "November": "11", "Desember": "12",
}

// parseTanggalIndo — "Minggu, 21 Agustus 2026" → "2026-08-21"
func parseTanggalIndo(tgl string) (string, bool) {
	parts := strings.Split(tgl, ", ")
	if len(parts) < 2 {
		return "", false
	}
	dayParts := strings.Split(parts[1], " ")
	if len(dayParts) < 3 {
		return "", false
	}
	dd := dayParts[0]
	if len(dd) == 1 {
		dd = "0" + dd
	}
	mm, ok := bulanMap[dayParts[1]]
	if !ok {
		return "", false
	}
	yyyy := dayParts[2]
	return fmt.Sprintf("%s-%s-%s", yyyy, mm, dd), true
}

// ScrapeResult — hasil scrape hari ini
type ScrapeResult struct {
	Success   bool
	JamMasuk  string
	JamPulang string
	Tanggal   string
	Status    string
	Source    string
	Error     string
}

// ScrapeToday — ambil presensi hari ini (WITA = UTC+8)
func (c *PusakaClient) ScrapeToday() ScrapeResult {
	now := time.Now()
	wita := now.Add(8 * time.Hour)
	bulan := int(wita.Month())
	tahun := wita.Year()
	tglISO := wita.Format("2006-01-02")

	riwayat, err := c.RiwayatPresensi(bulan, tahun)
	if err != nil {
		return ScrapeResult{Success: false, Tanggal: tglISO, Source: "http-api", Error: err.Error()}
	}

	for _, r := range riwayat {
		iso, ok := parseTanggalIndo(r.Tgl)
		if ok && iso == tglISO {
			// Tentukan status (reference + request user)
			status := determineStatus(r.JamMsk, r.JamPlg)
			return ScrapeResult{
				Success:   true,
				JamMasuk:  r.JamMsk,
				JamPulang: r.JamPlg,
				Tanggal:   tglISO,
				Status:    status,
				Source:    "http-api",
			}
		}
	}

	return ScrapeResult{Success: false, Tanggal: tglISO, Source: "http-api", Error: "Data hari ini tidak ditemukan"}
}

// determineStatus — logic status presensi (reference + Belum Masuk/Pulang)
func determineStatus(jamMasuk, jamPulang string) string {
	empty := func(s string) bool { return s == "" || s == "-" }
	if empty(jamMasuk) {
		return "Belum Masuk"
	}
	if empty(jamPulang) {
		return "Belum Pulang"
	}
	// Keduanya ada → banding jam masuk dengan 07:30 / 07:45 (reference)
	masuk := strings.ReplaceAll(jamMasuk, " ", "")
	masuk = strings.ReplaceAll(masuk, "WITA", "")
	masuk = masuk[:5] // "06:43"
	if masuk <= "07:30" {
		return "Tepat Waktu"
	} else if masuk <= "07:45" {
		return "Telat Ringan"
	}
	return "Terlambat"
}
