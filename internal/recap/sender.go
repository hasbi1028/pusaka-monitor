package recap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
)

// cfg — baca config GoWA dari env tiap panggil (bisa diubah tanpa restart)
func gowaBase() string  { return os.Getenv("GOWA_BASE_URL") }
func gowaUser() string  { return os.Getenv("GOWA_USER") }
func gowaPass() string  { return os.Getenv("GOWA_PASS") }
func gowaDev() string   { return firstNonEmpty(os.Getenv("GOWA_DEVICE"), "mtsnbot") }
// RecapGroup — target grup WA default dari env
func RecapGroup() string {
	return firstNonEmpty(os.Getenv("RECAP_GROUP"), "120363409303983377@g.us")
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

// SendImage — kirim PNG via GoWA /send/image (multipart) dengan compress=false.
func SendImage(to string, caption string, png []byte) error {
	base := gowaBase()
	if base == "" {
		return fmt.Errorf("GOWA_BASE_URL belum diset — pengiriman WA nonaktif")
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("phone", to); err != nil {
		return err
	}
	if err := w.WriteField("caption", caption); err != nil {
		return err
	}
	if err := w.WriteField("compress", "false"); err != nil {
		return err
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="rekap.png"`)
	h.Set("Content-Type", "image/png")
	fw, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := fw.Write(png); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", base+"/send/image", &body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(gowaUser(), gowaPass())
	req.Header.Set("X-Device-Id", gowaDev())
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GoWA %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(b, &out)
	if out.Code != "" && out.Code != "SUCCESS" {
		return fmt.Errorf("GoWA %s: %s", out.Code, out.Message)
	}
	return nil
}

// Result — hasil kirim satu instansi
type Result struct {
	InstansiID   string `json:"instansi_id"`
	InstansiNama string `json:"instansi_nama"`
	Target       string `json:"target"`
	ViaTelegram  bool   `json:"via_telegram,omitempty"`
	Sent         bool   `json:"sent"`
	Error        string `json:"error,omitempty"`
}

// targetFor — grup WA instansi, fallback ke RECAP_GROUP env
func targetFor(ins models.Instansi) string {
	if ins.WaGroup != "" {
		return ins.WaGroup
	}
	return RecapGroup()
}

// resolveTarget — resolve WA target with per-schedule override
func resolveTarget(ins models.Instansi, overrideGroup string) string {
	if overrideGroup != "" {
		return overrideGroup
	}
	return targetFor(ins)
}

// SendDailyRecaps — bangun + kirim gambar rekap semua instansi aktif ke WA.
// overrideGroup: per-schedule WA group override (optional, percents-sign)
func SendDailyRecaps(db *gorm.DB, tanggal string, overrideGroup ...string) []Result {
	list, err := ActiveInstansis(db)
	if err != nil {
		return []Result{{Sent: false, Error: "gagal baca instansi: " + err.Error()}}
	}
	var groupOverride string
	if len(overrideGroup) > 0 {
		groupOverride = overrideGroup[0]
	}
	var results []Result
	for _, ins := range list {
		if !ins.RecapAktif {
			log.Printf("[RECAP] skip %s: rekap dinonaktifkan", ins.Nama)
			continue
		}
		r := Result{InstansiID: ins.ID, InstansiNama: ins.Nama, Target: resolveTarget(ins, groupOverride)}
		d, err := Build(db, ins.ID, tanggal)
		if err != nil {
			r.Error = "gagal susun rekap: " + err.Error()
			results = append(results, r)
			continue
		}
		png, err := RenderPNG(d)
		if err != nil {
			r.Error = "gagal gambar: " + err.Error()
			results = append(results, r)
			continue
		}
		caption := CaptionFor(ins.Nama, d)
		if err := SendImage(r.Target, caption, png); err != nil {
			r.Error = "gagal kirim WA: " + err.Error()
			results = append(results, r)
			continue
		}
		r.Sent = true
		results = append(results, r)
		log.Printf("[RECAP] terkirim: %s → %s (%s)", ins.Nama, r.Target, tanggal)
	}
	return results
}

// SendDailyRecapsTelegram — kirim rekap semua instansi aktif ke Telegram
func SendDailyRecapsTelegram(db *gorm.DB, tanggal string) []Result {
	list, err := ActiveInstansis(db)
	if err != nil {
		return []Result{{Sent: false, Error: "gagal baca instansi: " + err.Error()}}
	}
	var results []Result
	for _, ins := range list {
		if !ins.RecapAktif {
			continue
		}
		r := Result{InstansiID: ins.ID, InstansiNama: ins.Nama, ViaTelegram: true}
		d, err := Build(db, ins.ID, tanggal)
		if err != nil {
			r.Error = "gagal susun rekap: " + err.Error()
			results = append(results, r)
			continue
		}
		png, err := RenderPNG(d)
		if err != nil {
			r.Error = "gagal gambar: " + err.Error()
			results = append(results, r)
			continue
		}
		caption := CaptionFor(ins.Nama, d)
		if err := SendTelegramPhoto(caption, png); err != nil {
			r.Error = "gagal kirim Telegram: " + err.Error()
			results = append(results, r)
			continue
		}
		r.Sent = true
		results = append(results, r)
		log.Printf("[RECAP-TG] terkirim: %s (%s)", ins.Nama, tanggal)
	}
	return results
}

// AfterScheduleBatch — dipanggil scheduler setelah batch auto-scrape dibuat.
// Menunggu antrean hari ini habis lalu kirim rekap berdasarkan per-schedule config.
func AfterScheduleBatch(db *gorm.DB, scheduleID, label, today string, sched *models.Schedule) {
	go func() {
		// Tunggu worker mulai ambil job (maks 5 mnt)
		time.Sleep(60 * time.Second)
		deadline := time.Now().Add(60 * time.Minute)
		for time.Now().Before(deadline) {
			var n int64
			midnight := today + "T00:00:00+08:00"
			t0, err := time.Parse(time.RFC3339, midnight)
			if err != nil {
				break
			}
			db.Session(&gorm.Session{NewDB: true}).Model(&models.Job{}).
				Where("status IN ('pending','running') AND created_at >= ?", t0).
				Count(&n)
			if n == 0 {
				break
			}
			time.Sleep(15 * time.Second)
		}
		key := models.RecapLog{ScheduleID: scheduleID, Tanggal: today}
		if err := db.Create(&key).Error; err != nil {
			log.Printf("[RECAP] '%s' %s sudah terkirim (skip ganda)", label, today)
			return
		}

		log.Printf("[RECAP] batch '%s' selesai → kirim rekap %s (wa=%v, tg=%v, group=%s)",
			label, today, sched.WAEnabled, sched.TelegramEnabled, sched.WAGroup)

		if sched.WAEnabled {
			results := SendDailyRecaps(db, today, sched.WAGroup)
			for _, r := range results {
				if !r.Sent {
					log.Printf("[RECAP-WA] GAGAL %s: %s", r.InstansiNama, r.Error)
				}
			}
		}

		if sched.TelegramEnabled {
			results := SendDailyRecapsTelegram(db, today)
			for _, r := range results {
				if !r.Sent {
					log.Printf("[RECAP-TG] GAGAL %s: %s", r.InstansiNama, r.Error)
				}
			}
		}
	}()
}
