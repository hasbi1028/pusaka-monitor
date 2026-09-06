package models

import "time"

// Instansi (tenant)
type Instansi struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	Nama           string     `gorm:"not null" json:"nama"`
	JnsInstansi    string     `gorm:"default:kua" json:"jns_instansi"`
	Kabupaten      string     `gorm:"not null" json:"kabupaten"`
	Provinsi       string     `gorm:"not null" json:"provinsi"`
	Alamat         string     `json:"alamat"`
	Telepon        string     `json:"telepon"`
	Email          string     `json:"email"`
	Aktif          bool       `gorm:"default:true" json:"aktif"`
	Status         string     `gorm:"default:pending" json:"status"`
	// Jam kerja normal
	JamMasuk       string     `gorm:"default:07:30" json:"jam_masuk"`
	JamPulang      string     `gorm:"default:16:00" json:"jam_pulang"`
	Toleransi      int        `gorm:"default:15" json:"toleransi"`
	// Jam kerja Ramadan
	JamMasukRam    string     `gorm:"default:07:00" json:"jam_masuk_ram"`
	JamPulangRam   string     `gorm:"default:15:30" json:"jam_pulang_ram"`
	ToleransiRam   int        `gorm:"default:15" json:"toleransi_ram"`
	ModeRamadan    bool       `gorm:"default:false" json:"mode_ramadan"`
	// Grup WA tujuan rekap gambar harian (JID grup ...@g.us atau nomor 628xx).
	// Kosong = pakai RECAP_GROUP dari env (fallback).
	WaGroup        string     `gorm:"default:''" json:"wa_group"`
	// RecapAktif — false = instansi ini dilewati saat kirim rekap harian.
	RecapAktif     bool       `gorm:"default:true" json:"recap_aktif"`
	ApprovedAt     *time.Time `json:"approved_at"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RejectReason   string     `json:"reject_reason"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// User (auth)
type User struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	InstansiID   string     `json:"instansi_id"`
	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         string     `gorm:"default:admin" json:"role"`
	Aktif        bool       `gorm:"default:true" json:"aktif"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Pegawai
type Pegawai struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	InstansiID     string    `gorm:"not null" json:"instansi_id"`
	NIP            string    `gorm:"not null" json:"nip"`
	Nama           string    `gorm:"not null" json:"nama"`
	Jabatan        string    `json:"jabatan"`
	Golongan       string    `json:"golongan"`
	PasswordPusaka string    `json:"password_pusaka"`
	Aktif          bool      `gorm:"default:true" json:"aktif"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Absensi
type Absensi struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	InstansiID string    `gorm:"not null" json:"instansi_id"`
	NIP        string    `gorm:"not null" json:"nip"`
	Nama       string    `gorm:"not null" json:"nama"`
	Tanggal    string    `gorm:"not null" json:"tanggal"`
	JamMasuk   string    `gorm:"default:-" json:"jam_masuk"`
	JamPulang  string    `gorm:"default:-" json:"jam_pulang"`
	Status     string    `gorm:"default:-" json:"status"`
	ScrapedAt  time.Time `json:"scraped_at"`
}

// Job
type Job struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	InstansiID   string     `gorm:"not null" json:"instansi_id"`
	EmployeeID   string     `gorm:"not null" json:"employee_id"`
	EmployeeName string     `gorm:"not null" json:"employee_name"`
	Status       string     `gorm:"default:pending" json:"status"`
	Attempts     int        `gorm:"default:0" json:"attempts"`
	MaxAttempts  int        `gorm:"default:3" json:"max_attempts"`
	Error        string     `json:"error"`
	Tanggal      string     `json:"tanggal"`
	JamMasuk     string     `json:"jam_masuk"`
	JamPulang    string     `json:"jam_pulang"`
	WorkerID     string     `json:"worker_id"`
	Progress     int        `gorm:"default:0" json:"progress"`
	TotalSteps   int        `gorm:"default:10" json:"total_steps"`
	StepLabel    string     `gorm:"default:''" json:"step_label"`
	CreatedAt    time.Time  `json:"created_at"`
	ClaimedAt    *time.Time `json:"claimed_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// Session
type Session struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Setting
type Setting struct {
	Key        string    `gorm:"primaryKey" json:"key"`
	Value      string    `gorm:"not null" json:"value"`
	InstansiID string    `gorm:"not null" json:"instansi_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ApprovalLog
type ApprovalLog struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	InstansiID string    `gorm:"not null" json:"instansi_id"`
	Action     string    `gorm:"not null" json:"action"`
	Actor      string    `gorm:"default:system" json:"actor"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// Cuti — data cuti pegawai
type Cuti struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	InstansiID    string    `gorm:"not null" json:"instansi_id"`
	NIP           string    `gorm:"not null" json:"nip"`
	Nama          string    `gorm:"not null" json:"nama"`
	TanggalMulai  string    `gorm:"not null" json:"tanggal_mulai"` // YYYY-MM-DD
	TanggalAkhir  string    `gorm:"not null" json:"tanggal_akhir"` // YYYY-MM-DD
	Keterangan    string    `json:"keterangan"`
	ApprovedBy    string    `json:"approved_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// Schedule — jadwal auto-scrape
type Schedule struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	Jam             int       `gorm:"not null" json:"jam"`        // 0-23
	Menit           int       `gorm:"not null" json:"menit"`      // 0-59
	Label           string    `gorm:"not null" json:"label"`      // e.g. "Pagi", "Sore"
	Aktif           bool      `gorm:"default:true" json:"aktif"`
	Mode            string    `gorm:"default:all" json:"mode"`    // all | belum_masuk | belum_pulang
	TelegramEnabled bool      `gorm:"default:false" json:"telegram_enabled"`
	WAEnabled       bool      `gorm:"default:true" json:"wa_enabled"`
	WAGroup         string    `gorm:"default:''" json:"wa_group"` // per-schedule WA group override
	LastRun         string    `gorm:"default:''" json:"last_run"` // YYYY-MM-DD terakhir jalan
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RecapLog — jejak rekap terkirim per jadwal per hari (anti-kirim-ganda)
type RecapLog struct {
	ScheduleID string    `gorm:"primaryKey" json:"schedule_id"`
	Tanggal    string    `gorm:"primaryKey" json:"tanggal"` // YYYY-MM-DD
	CreatedAt  time.Time `json:"created_at"`
}

// API Response
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Rekap
type RekapHarian struct {
	Total      int64 `json:"total"`
	Hadir      int64 `json:"hadir"`
	Terlambat  int64 `json:"terlambat"`
	TidakHadir int64 `json:"tidak_hadir"`
	BelumMasuk int64 `json:"belum_masuk"`
	BelumPulang int64 `json:"belum_pulang"`
}

// JobStats
type JobStats struct {
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Done      int64 `json:"done"`
	Failed    int64 `json:"failed"`
	Cancelled int64 `json:"cancelled"`
}
