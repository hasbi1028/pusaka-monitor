package recap

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// cutiSet — NIP pegawai yang sedang cuti pada tanggal (YYYY-MM-DD),
// beserta nama dari data cuti. Dipakai overlay agar cuti tampil "Cuti"
// bukan alfa, walau baris absensinya dibuat sebelum cuti diset.
func cutiSet(db *gorm.DB, instansiID, tanggal string) map[string]string {
	out := map[string]string{}
	var rows []models.Cuti
	q := db.Session(&gorm.Session{NewDB: true}).
		Where("tanggal_mulai <= ? AND tanggal_akhir >= ?", tanggal, tanggal)
	if instansiID != "" {
		q = q.Where("instansi_id = ?", instansiID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		out[r.NIP] = r.Nama
	}
	return out
}

// isLibur — true bila tanggal adalah hari libur mingguan instansi
// (settings key "libur_mingguan", default "0"=Minggu) ATAU terdaftar
// di tabel hari_libur (milik instansi / global "").
func isLibur(db *gorm.DB, instansiID, tanggal string) bool {
	t, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return false
	}
	val := ""
	var s models.Setting
	if instansiID != "" {
		if err := db.Session(&gorm.Session{NewDB: true}).
			Where("key = ? AND instansi_id = ?", "libur_mingguan", instansiID).
			First(&s).Error; err == nil {
			val = s.Value
		}
	}
	if val == "" {
		var g models.Setting
		if err := db.Session(&gorm.Session{NewDB: true}).
			Where("key = ? AND instansi_id = ?", "libur_mingguan", "global").
			First(&g).Error; err == nil {
			val = g.Value
		}
	}
	if val == "" {
		val = "0" // default: Minggu libur
	}
	for _, p := range strings.Split(val, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && n == int(t.Weekday()) {
			return true
		}
	}
	var n int64
	db.Session(&gorm.Session{NewDB: true}).Model(&models.HariLibur{}).
		Where("tanggal = ? AND (instansi_id = ? OR instansi_id = '')", tanggal, instansiID).
		Count(&n)
	return n > 0
}
