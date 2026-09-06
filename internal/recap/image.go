package recap

import (
	"bytes"
	"fmt"
	"image/color"
	"os"

	"github.com/fogleman/gg"
)

const imgW = 800

var (
	cWhite   = color.RGBA{255, 255, 255, 255}
	cInk     = color.RGBA{17, 24, 39, 255}
	cMuted   = color.RGBA{107, 114, 128, 255}
	cLine    = color.RGBA{229, 231, 235, 255}
	cHeader  = color.RGBA{30, 64, 175, 255} // biru tua
	cGreen   = color.RGBA{22, 163, 74, 255}
	cYellow  = color.RGBA{202, 138, 4, 255}
	cOrange  = color.RGBA{234, 88, 12, 255}
	cCyan    = color.RGBA{8, 145, 178, 255}
	cRed     = color.RGBA{220, 38, 38, 255}
	cGrayBg  = color.RGBA{249, 250, 251, 255}
	cChipBgG = color.RGBA{220, 252, 231, 255}
	cChipBgY = color.RGBA{254, 249, 195, 255}
	cChipBgR = color.RGBA{254, 226, 226, 255}
	cChipBgO = color.RGBA{255, 237, 213, 255}
	cChipBgC = color.RGBA{207, 250, 254, 255}
	cChipBgX = color.RGBA{243, 244, 246, 255}
)

func fontPath(bold bool) string {
	// Urutan: env → font bawaan repo (Linux/DomCloud) → font Windows (lokal)
	name := "assets/fonts/Arial.ttf"
	if bold {
		name = "assets/fonts/ArialBd.ttf"
	}
	if v := os.Getenv("RECAP_FONT"); v != "" {
		return v
	}
	if _, err := os.Stat(name); err == nil {
		return name
	}
	if bold {
		return `C:\Windows\Fonts\arialbd.ttf`
	}
	return `C:\Windows\Fonts\arial.ttf`
}

// scale — render 2x (1600px) agar teks tajam saat di-zoom di HP.
// PNG lossless; ketajaman ditentukan resolusi, bukan kompresi.
const scale = 2.0

// RenderPNG — gambar rekap harian satu instansi → bytes PNG
func RenderPNG(d *Data) ([]byte, error) {
	rowH := 30.0
	headerH := 150.0
	statH := 86.0
	tableHeadH := 32.0
	footH := 44.0
	h := headerH + statH + tableHeadH + rowH*float64(len(d.Rows)) + footH

	dc := gg.NewContext(int(float64(imgW)*scale), int(h*scale))
	dc.Scale(scale, scale)
	dc.SetColor(cWhite)
	dc.Clear()

	// Header band
	dc.SetColor(cHeader)
	dc.DrawRectangle(0, 0, imgW, headerH)
	dc.Fill()
	if err := loadFont(dc, true, 26); err != nil {
		return nil, err
	}
	dc.SetColor(cWhite)
	dc.DrawStringAnchored("REKAP PRESENSI HARIAN", imgW/2, 34, 0.5, 0.5)
	if err := loadFont(dc, true, 20); err != nil {
		return nil, err
	}
	dc.DrawStringAnchored(d.InstansiNama, imgW/2, 68, 0.5, 0.5)
	if err := loadFont(dc, false, 15); err != nil {
		return nil, err
	}
	dc.DrawStringAnchored(d.TanggalIndo+"  •  Masuk "+d.JamMasukStd+"  •  Pulang "+d.JamPulangStd, imgW/2, 98, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("Total %d pegawai", d.Total), imgW/2, 122, 0.5, 0.5)

	// Stat cards (6 kotak)
	stats := []struct {
		label string
		val   int
		col   color.RGBA
	}{
		{"HADIR", d.Hadir, cGreen},
		{"TELAT", d.Terlambat, cYellow},
		{"BLM MASUK", d.BelumMasuk, cOrange},
		{"BLM PULANG", d.BelumPulang, cCyan},
		{"ALFA", d.Alfa, cRed},
		{"TOTAL", d.Total, cInk},
	}
	y0 := headerH + 10.0
	cardW := (float64(imgW) - 20 - 5*8) / 6
	for i, s := range stats {
		x := 10 + float64(i)*(cardW+8)
		dc.SetColor(cWhite)
		dc.DrawRoundedRectangle(x, y0, cardW, statH-20, 8)
		dc.Fill()
		dc.SetColor(s.col)
		dc.DrawRectangle(x, y0+statH-26, 4, 6)
		dc.Fill()
		if err := loadFont(dc, true, 22); err != nil {
			return nil, err
		}
		dc.SetColor(s.col)
		dc.DrawStringAnchored(fmt.Sprintf("%d", s.val), x+cardW/2, y0+24, 0.5, 0.5)
		if err := loadFont(dc, false, 10); err != nil {
			return nil, err
		}
		dc.SetColor(cMuted)
		dc.DrawStringAnchored(s.label, x+cardW/2, y0+48, 0.5, 0.5)
	}

	// Table header
	ty := headerH + statH
	dc.SetColor(cInk)
	dc.DrawRectangle(0, ty, imgW, tableHeadH)
	dc.Fill()
	if err := loadFont(dc, true, 12); err != nil {
		return nil, err
	}
	dc.SetColor(cWhite)
	dc.DrawString("NO", 14, ty+21)
	dc.DrawString("NAMA", 56, ty+21)
	dc.DrawStringAnchored("MASUK", 560, ty+21, 0.5, 0)
	dc.DrawStringAnchored("PULANG", 650, ty+21, 0.5, 0)
	dc.DrawStringAnchored("STATUS", 735, ty+21, 0.5, 0)

	// Rows
	if err := loadFont(dc, false, 12); err != nil {
		return nil, err
	}
	for i, r := range d.Rows {
		y := ty + tableHeadH + rowH*float64(i)
		if i%2 == 1 {
			dc.SetColor(cGrayBg)
			dc.DrawRectangle(0, y, imgW, rowH)
			dc.Fill()
		}
		dc.SetColor(cMuted)
		dc.DrawString(fmt.Sprintf("%d", i+1), 14, y+20)
		dc.SetColor(cInk)
		name := r.Nama
		if len(name) > 34 {
			name = name[:33] + "…"
		}
		dc.DrawString(name, 56, y+20)
		dc.SetColor(cMuted)
		dc.DrawStringAnchored(r.JamMasuk, 560, y+20, 0.5, 0)
		dc.DrawStringAnchored(r.JamPulang, 650, y+20, 0.5, 0)
		// Status chip
		bg, fg := chipColor(r.Status)
		w, _ := dc.MeasureString(r.Status)
		cw := w + 14
		cx := 735 - cw/2
		dc.SetColor(bg)
		dc.DrawRoundedRectangle(cx, y+5, cw, 20, 10)
		dc.Fill()
		dc.SetColor(fg)
		dc.DrawStringAnchored(r.Status, 735, y+15, 0.5, 0.5)
		// garis bawah
		dc.SetColor(cLine)
		dc.DrawLine(0, y+rowH, imgW, y+rowH)
		dc.Stroke()
	}

	// Footer
	fy := ty + tableHeadH + rowH*float64(len(d.Rows))
	if err := loadFont(dc, false, 11); err != nil {
		return nil, err
	}
	dc.SetColor(cMuted)
	dc.DrawStringAnchored("HANYA MEMBACA data Pusaka • bukan membuat absen", imgW/2, fy+28, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func loadFont(dc *gg.Context, bold bool, size float64) error {
	return dc.LoadFontFace(fontPath(bold), size)
}

func chipColor(s string) (color.RGBA, color.RGBA) {
	switch s {
	case "Tepat Waktu", "Hadir":
		return cChipBgG, cGreen
	case "Telat Ringan":
		return cChipBgY, cYellow
	case "Terlambat":
		return cChipBgR, cRed
	case "Belum Pulang":
		return cChipBgC, cCyan
	case "Belum Masuk":
		return cChipBgO, cOrange
	default:
		return cChipBgX, cMuted
	}
}
