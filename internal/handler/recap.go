package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hasbiawal/pusaka-monitor/internal/models"
	"github.com/hasbiawal/pusaka-monitor/internal/recap"
)

type RecapHandler struct {
	DB *gorm.DB
}

func (h *RecapHandler) mustSuperadmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	if role != "superadmin" {
		c.JSON(http.StatusForbidden, models.ApiResponse{Error: "khusus superadmin"})
		return false
	}
	return true
}

// GET /api/admin/recap/preview?instansi_id=&tanggal=YYYY-MM-DD → PNG
func (h *RecapHandler) Preview(c *gin.Context) {
	if !h.mustSuperadmin(c) {
		return
	}
	instansiID := c.Query("instansi_id")
	tanggal := c.DefaultQuery("tanggal", todayWITA())
	if instansiID == "" {
		// default: instansi pertama
		list, err := recap.ActiveInstansis(h.DB)
		if err != nil || len(list) == 0 {
			c.JSON(http.StatusNotFound, models.ApiResponse{Error: "tidak ada instansi aktif"})
			return
		}
		instansiID = list[0].ID
	}
	d, err := recap.Build(h.DB, instansiID, tanggal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal susun rekap: " + err.Error()})
		return
	}
	png, err := recap.RenderPNG(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal gambar: " + err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}

// POST /api/admin/recap/send?tanggal=YYYY-MM-DD → kirim ke grup WA sekarang
func (h *RecapHandler) SendNow(c *gin.Context) {
	if !h.mustSuperadmin(c) {
		return
	}
	tanggal := c.DefaultQuery("tanggal", todayWITA())
	results := recap.SendDailyRecaps(h.DB, tanggal)
	ok := 0
	for _, r := range results {
		if r.Sent {
			ok++
		}
	}
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: ok > 0,
		Data:    gin.H{"tanggal": tanggal, "results": results},
	})
}

// kirimAllResult — helper hasil kirim semua instansi (superadmin tanpa instansi_id)
func (h *RecapHandler) kirimAll(c *gin.Context, via string, tanggal string) {
	var results []recap.Result
	if via == "wa" {
		results = recap.SendDailyRecaps(h.DB, tanggal)
	} else {
		list, err := recap.ActiveInstansis(h.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal baca instansi: " + err.Error()})
			return
		}
		if len(list) == 0 {
			c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "tidak ada instansi aktif"})
			return
		}
		for _, ins := range list {
			r := recap.Result{InstansiID: ins.ID, InstansiNama: ins.Nama}
			if !ins.RecapAktif {
				r.Error = "rekap dinonaktifkan"
				results = append(results, r)
				continue
			}
			d, err := recap.Build(h.DB, ins.ID, tanggal)
			if err != nil {
				r.Error = "gagal susun rekap: " + err.Error()
				results = append(results, r)
				continue
			}
			png, err := recap.RenderPNG(d)
			if err != nil {
				r.Error = "gagal gambar: " + err.Error()
				results = append(results, r)
				continue
			}
			caption := recap.CaptionFor(ins.Nama, d)
			if err := recap.SendTelegramPhoto(caption, png); err != nil {
				r.Error = "gagal kirim Telegram: " + err.Error()
				results = append(results, r)
				continue
			}
			r.Sent = true
			results = append(results, r)
			log.Printf("[RECAP-TG] manual kirim (all): %s (%s)", ins.Nama, tanggal)
		}
	}
	ok := 0
	for _, r := range results {
		if r.Sent {
			ok++
		}
	}
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: ok > 0,
		Data:    gin.H{"tanggal": tanggal, "via": via, "ok": ok, "total": len(results), "results": results},
	})
}

// POST /api/rekap/kirim-wa — kirim rekap instansi tertentu ke grup WA
func (h *RecapHandler) SendWA(c *gin.Context) {
	role, _ := c.Get("role")
	instansiID, _ := c.Get("instansi_id")
	tanggal := c.DefaultQuery("tanggal", todayWITA())

	targetID, _ := instansiID.(string)
	if role == "superadmin" {
		if q := c.Query("instansi_id"); q != "" {
			targetID = q
		}
	}
	if targetID == "" {
		// Superadmin tanpa instansi → kirim semua instansi aktif
		h.kirimAll(c, "wa", tanggal)
		return
	}
	var ins models.Instansi
	if err := h.DB.First(&ins, "id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "instansi tidak ditemukan"})
		return
	}
	if !ins.RecapAktif {
		c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "rekap instansi ini dinonaktifkan"})
		return
	}
	d, err := recap.Build(h.DB, targetID, tanggal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal susun rekap: " + err.Error()})
		return
	}
	png, err := recap.RenderPNG(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal gambar: " + err.Error()})
		return
	}
	target := ins.WaGroup
	if target == "" {
		target = recap.RecapGroup()
	}
	if target == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "target grup WA belum diatur (RECAP_GROUP atau wa_group instansi)"})
		return
	}
	caption := recap.CaptionFor(ins.Nama, d)
	if err := recap.SendImage(target, caption, png); err != nil {
		c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "gagal kirim WA: " + err.Error()})
		return
	}
	log.Printf("[RECAP-WA] manual kirim: %s → %s (%s)", ins.Nama, target, tanggal)
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"tanggal": tanggal, "target": target, "instansi": ins.Nama, "via": "wa"},
	})
}

// POST /api/rekap/kirim-telegram — kirim rekap instansi tertentu ke Telegram
func (h *RecapHandler) SendTelegram(c *gin.Context) {
	role, _ := c.Get("role")
	instansiID, _ := c.Get("instansi_id")
	tanggal := c.DefaultQuery("tanggal", todayWITA())

	targetID, _ := instansiID.(string)
	if role == "superadmin" {
		if q := c.Query("instansi_id"); q != "" {
			targetID = q
		}
	}
	if targetID == "" {
		// Superadmin tanpa instansi → kirim semua instansi aktif
		if !recap.TelegramReady() {
			c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID belum diset"})
			return
		}
		h.kirimAll(c, "telegram", tanggal)
		return
	}
	var ins models.Instansi
	if err := h.DB.First(&ins, "id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ApiResponse{Error: "instansi tidak ditemukan"})
		return
	}
	d, err := recap.Build(h.DB, targetID, tanggal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal susun rekap: " + err.Error()})
		return
	}
	png, err := recap.RenderPNG(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal gambar: " + err.Error()})
		return
	}
	caption := recap.CaptionFor(ins.Nama, d)
	if err := recap.SendTelegramPhoto(caption, png); err != nil {
		c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "gagal kirim Telegram: " + err.Error()})
		return
	}
	log.Printf("[RECAP-TG] manual kirim: %s (%s)", ins.Nama, tanggal)
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"tanggal": tanggal, "instansi": ins.Nama, "via": "telegram"},
	})
}

// POST /api/recap/test → admin instansi kirim tes rekap instansinya sendiri
func (h *RecapHandler) TestSend(c *gin.Context) {
	instansiID, _ := c.Get("instansi_id")
	role, _ := c.Get("role")
	tanggal := c.DefaultQuery("tanggal", todayWITA())

	targetID, _ := instansiID.(string)
	if role == "superadmin" {
		if q := c.Query("instansi_id"); q != "" {
			targetID = q
		} else {
			// superadmin tanpa parameter = kirim semua (sama seperti SendNow)
			results := recap.SendDailyRecaps(h.DB, tanggal)
			c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: gin.H{"tanggal": tanggal, "results": results}})
			return
		}
	}
	if targetID == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "instansi tidak dikenali"})
		return
	}
	d, err := recap.Build(h.DB, targetID, tanggal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal susun rekap: " + err.Error()})
		return
	}
	var ins models.Instansi
	h.DB.First(&ins, "id = ?", targetID)
	if !ins.RecapAktif {
		c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "rekap instansi ini dinonaktifkan"})
		return
	}
	target := ins.WaGroup
	if target == "" {
		target = os.Getenv("RECAP_GROUP")
		if target == "" {
			target = "120363409303983377@g.us"
		}
	}
	png, err := recap.RenderPNG(d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "gagal gambar: " + err.Error()})
		return
	}
	caption := "[TES] " + recap.CaptionFor(ins.Nama, d)
	if err := recap.SendImage(target, caption, png); err != nil {
		c.JSON(http.StatusOK, models.ApiResponse{Success: false, Error: "gagal kirim WA: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    gin.H{"tanggal": tanggal, "target": target, "instansi": ins.Nama},
	})
}

// GET /api/admin/wa/groups — fetch WA groups from GoWA
func (h *RecapHandler) ListWAGroups(c *gin.Context) {
	if !h.mustSuperadmin(c) {
		return
	}
	base := os.Getenv("GOWA_BASE_URL")
	user := os.Getenv("GOWA_USER")
	pass := os.Getenv("GOWA_PASS")
	device := os.Getenv("GOWA_DEVICE")
	if device == "" {
		device = "mtsnbot"
	}
	if base == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{Error: "GOWA_BASE_URL belum diset"})
		return
	}

	type groupInfo struct {
		ID       string `json:"id"`
		Subject  string `json:"subject"`
		Name     string `json:"name"`
	}

	req, err := http.NewRequest("GET", base+"/group/list", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal membuat request"})
		return
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("X-Device-Id", device)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ApiResponse{Error: "Gagal koneksi GoWA: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // max 1MB

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, models.ApiResponse{Error: fmt.Sprintf("GoWA %d: %s", resp.StatusCode, string(body[:min(len(body), 500)]))})
		return
	}

	// Try to parse as array of groups
	var raw []struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(body, &raw); err == nil {
		groups := make([]gin.H, 0, len(raw))
		for _, g := range raw {
			name := g.Subject
			if name == "" {
				name = g.Name
			}
			groups = append(groups, gin.H{"id": g.ID, "name": name})
		}
		c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: groups})
		return
	}

	// Try to parse as object with groups array
	var wrapper struct {
		Groups []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Name    string `json:"name"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(body, &wrapper); err == nil && len(wrapper.Groups) > 0 {
		groups := make([]gin.H, 0, len(wrapper.Groups))
		for _, g := range wrapper.Groups {
			name := g.Subject
			if name == "" {
				name = g.Name
			}
			groups = append(groups, gin.H{"id": g.ID, "name": name})
		}
		c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: groups})
		return
	}

	// Raw response — return as-is
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: json.RawMessage(body)})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
