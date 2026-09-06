package recap

import (
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// Row — satu baris pegawai di rekap
type Row struct {
	Nama      string
	JamMasuk  string
	JamPulang string
	Status    string
}

// Data — rekap harian satu instansi
type Data struct {
	InstansiID   string
	InstansiNama string
	Kabupaten    string
	Tanggal      string // YYYY-MM-DD
	TanggalIndo  string // "Kamis, 3 September 2026"
	JamMasukStd  string
	JamPulangStd string
	Total        int
	Hadir        int
	Terlambat    int
	BelumMasuk   int
	BelumPulang  int
	Alfa         int
	Cuti         int
	Libur        int
	IsLibur      bool // true bila tanggal adalah hari libur
	Rows         []Row
}

// Build — susun data rekap harian satu instansi (query fresh, NewDB)
func Build(db *gorm.DB, instansiID, tanggal string) (*Data, error) {
	var ins models.Instansi
	if err := db.Session(&gorm.Session{NewDB: true}).
		Where("id = ?", instansiID).First(&ins).Error; err != nil {
		return nil, err
	}

	var list []models.Absensi
	if err := db.Session(&gorm.Session{NewDB: true}).
		Where("instansi_id = ? AND tanggal = ?", instansiID, tanggal).
		Order("nama").Find(&list).Error; err != nil {
		return nil, err
	}

	d := &Data{
		InstansiID:   ins.ID,
		InstansiNama: ins.Nama,
		Kabupaten:    ins.Kabupaten,
		Tanggal:      tanggal,
		TanggalIndo:  formatIndo(tanggal),
		JamMasukStd:  ins.JamMasuk,
		JamPulangStd: ins.JamPulang,
		Total:        len(list),
	}
	// Resolve status: Libur > Cuti > hasil scrape (sama dengan dashboard).
	libur := isLibur(db, instansiID, tanggal)
	d.IsLibur = libur
	cuti := map[string]string{}
	if !libur {
		cuti = cutiSet(db, instansiID, tanggal)
	}
	empty := func(s string) bool { return s == "" || s == "-" }
	for _, a := range list {
		if libur {
			a.Status = "Libur"
			d.Libur++
			d.Rows = append(d.Rows, Row{
				Nama: a.Nama, JamMasuk: orDash(a.JamMasuk),
				JamPulang: orDash(a.JamPulang), Status: a.Status,
			})
			continue
		}
		if _, ok := cuti[a.NIP]; ok {
			a.Status = "Cuti"
			d.Cuti++
			d.Rows = append(d.Rows, Row{
				Nama: a.Nama, JamMasuk: orDash(a.JamMasuk),
				JamPulang: orDash(a.JamPulang), Status: a.Status,
			})
			continue
		}
		masuk := empty(a.JamMasuk)
		pulang := empty(a.JamPulang)
		switch {
		case masuk && pulang:
			d.Alfa++
		case a.Status == "Terlambat" || a.Status == "Telat Ringan":
			d.Terlambat++
			d.Hadir++
		case a.Status == "Belum Masuk":
			d.BelumMasuk++
		case a.Status == "Belum Pulang":
			d.BelumPulang++
			d.Hadir++
		default:
			d.Hadir++
		}
		if a.Status == "" {
			if !masuk {
				a.Status = "Hadir"
			} else {
				a.Status = "Alfa"
			}
		}
		d.Rows = append(d.Rows, Row{
			Nama: a.Nama, JamMasuk: orDash(a.JamMasuk),
			JamPulang: orDash(a.JamPulang), Status: a.Status,
		})
	}
	return d, nil
}

// ActiveInstansis — daftar instansi aktif
func ActiveInstansis(db *gorm.DB) ([]models.Instansi, error) {
	var list []models.Instansi
	err := db.Session(&gorm.Session{NewDB: true}).
		Where("aktif = ? AND status = ?", true, "approved").
		Order("nama").Find(&list).Error
	return list, err
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

var namaHari = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
var namaBulan = []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

// formatIndo — "2026-09-03" → "Kamis, 3 September 2026"
func formatIndo(tgl string) string {
	if len(tgl) != 10 {
		return tgl
	}
	y := atoi(tgl[0:4])
	m := atoi(tgl[5:7])
	d := atoi(tgl[8:10])
	if m < 1 || m > 12 {
		return tgl
	}
	// Zeller untuk nama hari (Gregorian)
	q := d
	mm := m
	yy := y
	if mm < 3 {
		mm += 12
		yy--
	}
	k := yy % 100
	j := yy / 100
	h := (q + (13*(mm+1))/5 + k + k/4 + j/4 + 5*j) % 7 // 0=Sabtu
	hari := (h + 6) % 7                               // 0=Minggu
	return namaHari[hari] + ", " + itoa(d) + " " + namaBulan[m] + " " + itoa(y)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
