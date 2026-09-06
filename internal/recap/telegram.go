package recap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"
)

func tgToken() string { return os.Getenv("TELEGRAM_BOT_TOKEN") }
func tgChat() string  { return os.Getenv("TELEGRAM_CHAT_ID") }

// TelegramReady — true bila token + chat terisi
func TelegramReady() bool { return tgToken() != "" && tgChat() != "" }

// SendTelegramPhoto — kirim PNG via Bot API sendPhoto (fallback bila WA gagal)
func SendTelegramPhoto(caption string, png []byte) error {
	token := tgToken()
	chat := tgChat()
	if token == "" || chat == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID belum diset")
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("chat_id", chat); err != nil {
		return err
	}
	if err := w.WriteField("caption", caption); err != nil {
		return err
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="photo"; filename="rekap.png"`)
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

	req, err := http.NewRequest("POST",
		"https://api.telegram.org/bot"+token+"/sendPhoto", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(b, &out)
	if !out.OK {
		return fmt.Errorf("telegram: %s", out.Description)
	}
	return nil
}
