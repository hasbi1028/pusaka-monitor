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
			return ScrapeResult{
				Success:   true,
				JamMasuk:  r.JamMsk,
				JamPulang: r.JamPlg,
				Tanggal:   tglISO,
				Source:    "http-api",
			}
		}
	}

	return ScrapeResult{Success: false, Tanggal: tglISO, Source: "http-api", Error: "Data hari ini tidak ditemukan"}
}

// DetermineStatus — logic status presensi berdasarkan jam masuk instansi
// jamMasukStd: "07:30", toleransi: 15 menit, jamPulangStd: "16:00"
func DetermineStatus(jamMasuk, jamPulang, jamMasukStd string, toleransi int) string {
	empty := func(s string) bool { return s == "" || s == "-" }
	if empty(jamMasuk) {
		return "Belum Masuk"
	}
	if empty(jamPulang) {
		return "Belum Pulang"
	}
	// Keduanya ada → banding jam masuk dengan standar
	masuk := strings.ReplaceAll(jamMasuk, " ", "")
	masuk = strings.ReplaceAll(masuk, "WITA", "")
	if len(masuk) >= 5 {
		masuk = masuk[:5] // "06:43"
	}

	// Hitung batas telat ringan = jamMasukStd + toleransi menit
	batasTelat := addMinutes(jamMasukStd, toleransi)

	if masuk <= jamMasukStd {
		return "Tepat Waktu"
	} else if masuk <= batasTelat {
		return "Telat Ringan"
	}
	return "Terlambat"
}

// addMinutes — "07:30" + 15 → "07:45"
func addMinutes(jamStr string, menit int) string {
	parts := strings.Split(jamStr, ":")
	if len(parts) != 2 {
		return jamStr
	}
	h := parseInt(parts[0])
	m := parseInt(parts[1])
	m += menit
	for m >= 60 {
		h++
		m -= 60
	}
	for m < 0 {
		h--
		m += 60
	}
	if h > 23 {
		h = 23
	}
	if h < 0 {
		h = 0
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
