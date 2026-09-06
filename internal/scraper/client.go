package scraper

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Konstanta Pusaka v3 (reverse-engineered dari worker lama)
const (
	V3Base  = "https://pusaka-v3.kemenag.go.id"
	AuthBase = "https://pusaka-auth.kemenag.go.id"

	DefaultUA       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
	TokenTrustMins  = 50
	RequestTimeout  = 90 * time.Second
	MinDelayMs      = 5000
	MaxDelayMs      = 12000
	Backoff429Mins  = 30
	Backoff403Mins  = 60
	MaxReqPerHour   = 60
)

// Anti-ban: rate limiter PER-CLIENT (per-NIP).
// Tiap PusakaClient (1 NIP) punya kuota 60 req/jam sendiri,
// sehingga 35 NIP berbeda bisa jalan paralel penuh tanpa
// saling antre (berbeda dengan limiter global).
type RiwayatDay struct {
	Tgl       string `json:"tgl"`
	JamMsk    string `json:"jam_msk"`
	JamPlg    string `json:"jam_plg"`
	TzMsk     string `json:"tz_msk"`
	TzPlg     string `json:"tz_plg"`
	KetLibur  string `json:"keterangan_libur"`
}

type riwayatResponse struct {
	Status bool         `json:"status"`
	Data   []RiwayatDay `json:"data"`
}

// savedSession — struktur persist session ke file
type savedSession struct {
	Cookies      map[string]string `json:"cookies"`
	Token        string            `json:"token"`
	TokenSavedAt time.Time         `json:"token_saved_at"`
}

// TotalSteps — jumlah langkah scrape (ditampilkan sebagai n/10 + persen)
const TotalSteps = 10

// ProgressFunc — callback tiap langkah scrape (step 1..TotalSteps + label)
type ProgressFunc func(step int, label string)

// PusakaClient — klien HTTP ke Pusaka v3 (NextAuth flow)
type PusakaClient struct {
	NIP         string
	Nama        string
	Password    string
	OnProgress  ProgressFunc
	cookies     map[string]string
	token       string
	tokenSavedAt time.Time
	mu          sync.Mutex
}

// report — panggil callback progres (aman dari nil)
func (c *PusakaClient) report(step int, label string) {
	if c.OnProgress != nil {
		c.OnProgress(step, label)
	}
}

// NewClient — factory
func NewClient(nip, nama, password string) *PusakaClient {
	c := &PusakaClient{
		NIP:      nip,
		Nama:     nama,
		Password: password,
		cookies:  map[string]string{},
	}
	c.loadSession() // reuse session dari file kalau masih fresh
	return c
}

// sessionFile — path file session per NIP (hash sha256 12 char)
func sessionFile(nip string) string {
	h := sha256.Sum256([]byte(nip))
	name := fmt.Sprintf("%x", h)[:12]
	_ = os.MkdirAll("data/sessions", 0755)
	return filepath.Join("data/sessions", name+".json")
}

func (c *PusakaClient) loadSession() {
	data, err := os.ReadFile(sessionFile(c.NIP))
	if err != nil {
		return
	}
	var s savedSession
	if err := json.Unmarshal(data, &s); err != nil {
		return
	}
	c.mu.Lock()
	c.cookies = s.Cookies
	if s.Cookies == nil {
		c.cookies = map[string]string{}
	}
	c.token = s.Token
	c.tokenSavedAt = s.TokenSavedAt
	c.mu.Unlock()
}

func (c *PusakaClient) saveSession() {
	c.mu.Lock()
	s := savedSession{
		Cookies:      c.cookies,
		Token:        c.token,
		TokenSavedAt: c.tokenSavedAt,
	}
	c.mu.Unlock()
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(sessionFile(c.NIP), data, 0644)
}

// antiBanDelay — jeda kecil antar request dalam 1 sesi login
// (reference: jitter 500-1500ms, BUKAN 60dtk)
func (c *PusakaClient) antiBanDelay() {
	jitterSleep(500, 1500)
}

func jitterSleep(min, max int) {
	ms := min + rand.Intn(max-min)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// doRequest — fetch dengan cookie jar + anti-ban headers
func (c *PusakaClient) doRequest(method, rawURL, body, bearer string, form bool) (*http.Response, error) {
	req, err := http.NewRequest(method, rawURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en;q=0.8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	c.mu.Lock()
	if len(c.cookies) > 0 {
		var parts []string
		for k, v := range c.cookies {
			parts = append(parts, k+"="+v)
		}
		req.Header.Set("Cookie", strings.Join(parts, "; "))
	}
	c.mu.Unlock()

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if body != "" {
		if form {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	client := &http.Client{
		Timeout: RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // manual redirect (anti-ban)
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Absorb set-cookies
	for _, ck := range resp.Cookies() {
		c.mu.Lock()
		c.cookies[ck.Name] = ck.Value
		c.mu.Unlock()
	}
	return resp, nil
}

func assertNotBlocked(status int, ctx string) error {
	switch status {
	case 429:
		return fmt.Errorf("terblokir WAF (%s → 429 rate limit, backoff %d menit)", ctx, Backoff429Mins)
	case 403:
		return fmt.Errorf("terblokir WAF (%s → 403 block, backoff %d menit)", ctx, Backoff403Mins)
	case 503:
		return fmt.Errorf("terblokir WAF (%s → 503)", ctx)
	}
	return nil
}

// Login — NextAuth credentials flow
func (c *PusakaClient) Login() (string, error) {
	// Rate limit per-client
	c.antiBanDelay()

	// 1. CSRF
	c.report(3, "Ambil CSRF")
	csrfResp, err := c.doRequest("GET", V3Base+"/api/auth/csrf", "", "", false)
	if err != nil {
		return "", err
	}
	if err := assertNotBlocked(csrfResp.StatusCode, "GET /api/auth/csrf"); err != nil {
		return "", err
	}
	var csrfBody struct {
		CsrfToken string `json:"csrfToken"`
	}
	json.NewDecoder(csrfResp.Body).Decode(&csrfBody)
	csrfResp.Body.Close()
	if csrfBody.CsrfToken == "" {
		return "", fmt.Errorf("csrfToken tidak ditemukan")
	}

	jitterSleep(500, 1500)

	// 2. POST credentials
	c.report(4, "Login Pusaka")
	form := url.Values{}
	form.Set("csrfToken", csrfBody.CsrfToken)
	form.Set("email", c.NIP)
	form.Set("password", c.Password)
	form.Set("json", "true")

	c.antiBanDelay()
	loginResp, err := c.doRequest("POST", V3Base+"/api/auth/callback/credentials", form.Encode(), "", true)
	if err != nil {
		return "", err
	}
	loginResp.Body.Close()
	if err := assertNotBlocked(loginResp.StatusCode, "POST callback/credentials"); err != nil {
		return "", err
	}
	if loginResp.StatusCode == 401 {
		return "", fmt.Errorf("kredensial ditolak (401)")
	}

	jitterSleep(500, 1000)

	// 3. GET token
	c.report(5, "Ambil token sesi")
	c.antiBanDelay()
	tokenResp, err := c.doRequest("GET", V3Base+"/api/auth/session", "", "", false)
	if err != nil {
		return "", err
	}
	if err := assertNotBlocked(tokenResp.StatusCode, "GET /api/auth/session"); err != nil {
		return "", err
	}
	var tokenBody struct {
		Token string `json:"token"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenBody)
	tokenResp.Body.Close()

	if tokenBody.Token == "" {
		return "", fmt.Errorf("session tidak menghasilkan token")
	}
	c.token = tokenBody.Token
	c.tokenSavedAt = time.Now()
	c.saveSession() // persist biar reuse
	return c.token, nil
}

// EnsureToken — reuse jika masih fresh (<50 menit) dari file session
func (c *PusakaClient) EnsureToken() (string, error) {
	c.report(2, "Cek sesi")
	c.mu.Lock()
	token := c.token
	savedAt := c.tokenSavedAt
	c.mu.Unlock()
	if token != "" {
		if time.Since(savedAt) < TokenTrustMins*time.Minute {
			return token, nil
		}
	}
	return c.Login()
}

// RiwayatPresensi — GET riwayat per bulan/tahun
func (c *PusakaClient) RiwayatPresensi(bulan, tahun int) ([]RiwayatDay, error) {
	token, err := c.EnsureToken()
	if err != nil {
		return nil, err
	}
	c.report(6, "Ambil riwayat")
	c.antiBanDelay()

	u := fmt.Sprintf("%s/presensi/api/riwayat-presensi?bulan=%d&tahun=%d", AuthBase, bulan, tahun)
	resp, err := c.doRequest("GET", u, "", token, false)
	if err != nil {
		return nil, err
	}
	if err := assertNotBlocked(resp.StatusCode, "GET riwayat-presensi"); err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
		newToken, e := c.Login()
		if e != nil {
			return nil, e
		}
		jitterSleep(500, 1500)
		resp2, e2 := c.doRequest("GET", u, "", newToken, false)
		if e2 != nil {
			return nil, e2
		}
		resp = resp2
	}
	defer resp.Body.Close()

	var rb riwayatResponse
	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		return nil, err
	}
	return rb.Data, nil
}
