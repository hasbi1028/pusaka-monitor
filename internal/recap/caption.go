package recap

import (
	"fmt"
	"strings"
)

// CaptionFor — teks rekap untuk WA/Telegram, selalu konsisten dengan gambar.
// Pegawai cuti disebut eksplisit (nama) agar jelas bukan alfa.
func CaptionFor(instansiNama string, d *Data) string {
	cap := fmt.Sprintf("Rekap %s\n%s\nHadir %d • Telat %d • Blm Masuk %d • Blm Pulang %d • Alfa %d • Cuti %d",
		instansiNama, d.TanggalIndo, d.Hadir, d.Terlambat, d.BelumMasuk, d.BelumPulang, d.Alfa, d.Cuti)
	if d.IsLibur {
		cap += "\nHari libur — tidak ada auto-scrape"
		return cap
	}
	if d.Cuti > 0 {
		names := []string{}
		for _, r := range d.Rows {
			if r.Status == "Cuti" {
				names = append(names, r.Nama)
			}
		}
		s := strings.Join(names, ", ")
		if len(s) > 150 {
			s = s[:147] + "..."
		}
		cap += "\nCuti: " + s
	}
	return cap
}
