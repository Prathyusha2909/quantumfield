package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleAdmin   = "admin"
	RoleAnalyst = "analyst"

	ScanQueued    = "queued"
	ScanRunning   = "running"
	ScanCompleted = "completed"
	ScanFailed    = "failed"

	AuditLoginSuccess   = "LOGIN_SUCCESS"
	AuditLoginFailed    = "LOGIN_FAILED"
	AuditUserRegistered = "USER_REGISTERED"
	AuditAssetCreated   = "ASSET_CREATED"
	AuditScanQueued     = "SCAN_QUEUED"
	AuditScanCompleted  = "SCAN_COMPLETED"
	AuditScanFailed     = "SCAN_FAILED"
	AuditReportExported = "REPORT_EXPORTED"
)

type Base struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (base *Base) EnsureID() {
	if base.ID == "" {
		base.ID = uuid.NewString()
	}
}

type User struct {
	Base
	Name         string  `json:"name" gorm:"size:120;not null"`
	Email        string  `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string  `json:"-" gorm:"not null"`
	Role         string  `json:"role" gorm:"size:32;not null;default:analyst"`
	Assets       []Asset `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	user.EnsureID()
	return nil
}

type Asset struct {
	Base
	UserID           string     `json:"user_id" gorm:"type:uuid;not null;index"`
	Domain           string     `json:"domain" gorm:"size:253;not null;index:idx_user_domain,unique"`
	Port             int        `json:"port" gorm:"not null;default:443;index:idx_user_domain,unique"`
	Label            string     `json:"label" gorm:"size:120"`
	Status           string     `json:"status" gorm:"size:32;not null;default:pending"`
	LastScannedAt    *time.Time `json:"last_scanned_at"`
	CurrentRiskScore int        `json:"current_risk_score"`
	CurrentPQCScore  int        `json:"current_pqc_score"`
	Scans            []Scan     `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

func (asset *Asset) BeforeCreate(_ *gorm.DB) error {
	asset.EnsureID()
	return nil
}

type Scan struct {
	Base
	AssetID       string         `json:"asset_id" gorm:"type:uuid;not null;index"`
	Asset         *Asset         `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
	Status        string         `json:"status" gorm:"size:32;not null;index"`
	StartedAt     *time.Time     `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	DurationMS    int64          `json:"duration_ms"`
	RiskScore     int            `json:"risk_score"`
	PQCScore      int            `json:"pqc_score"`
	TLSVersion    string         `json:"tls_version" gorm:"size:32"`
	CipherSuite   string         `json:"cipher_suite" gorm:"size:128"`
	ErrorMessage  string         `json:"error_message,omitempty" gorm:"type:text"`
	RetryCount    int            `json:"retry_count" gorm:"not null;default:0"`
	MaxRetries    int            `json:"max_retries" gorm:"not null;default:3"`
	LastError     string         `json:"last_error,omitempty" gorm:"type:text"`
	FailedAt      *time.Time     `json:"failed_at"`
	Certificate   *Certificate   `json:"certificate,omitempty"`
	Findings      []Finding      `json:"findings,omitempty"`
	PQCAssessment *PQCAssessment `json:"pqc_assessment,omitempty"`
}

func (scan *Scan) BeforeCreate(_ *gorm.DB) error {
	scan.EnsureID()
	return nil
}

type Certificate struct {
	Base
	ScanID             string    `json:"scan_id" gorm:"type:uuid;not null;uniqueIndex"`
	AssetID            string    `json:"asset_id" gorm:"type:uuid;not null;index"`
	Asset              *Asset    `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
	Subject            string    `json:"subject" gorm:"type:text"`
	CommonName         string    `json:"common_name" gorm:"size:255"`
	Issuer             string    `json:"issuer" gorm:"type:text"`
	SerialNumber       string    `json:"serial_number" gorm:"size:255"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after" gorm:"index"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm" gorm:"size:64"`
	SignatureAlgorithm string    `json:"signature_algorithm" gorm:"size:128"`
	KeySize            int       `json:"key_size"`
	ChainValid         bool      `json:"chain_valid"`
	HostnameValid      bool      `json:"hostname_valid"`
	SelfSigned         bool      `json:"self_signed"`
	SubjectAltNames    []string  `json:"subject_alt_names" gorm:"serializer:json"`
	FingerprintSHA256  string    `json:"fingerprint_sha256" gorm:"size:95"`
	PEM                string    `json:"pem,omitempty" gorm:"type:text"`
}

func (certificate *Certificate) BeforeCreate(_ *gorm.DB) error {
	certificate.EnsureID()
	return nil
}

type Finding struct {
	Base
	ScanID      string `json:"scan_id" gorm:"type:uuid;not null;index"`
	AssetID     string `json:"asset_id" gorm:"type:uuid;not null;index"`
	Asset       *Asset `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
	Type        string `json:"type" gorm:"size:80;not null;index"`
	Severity    string `json:"severity" gorm:"size:16;not null;index"`
	Title       string `json:"title" gorm:"size:180;not null"`
	Description string `json:"description" gorm:"type:text"`
	Remediation string `json:"remediation" gorm:"type:text"`
	Evidence    string `json:"evidence" gorm:"type:text"`
	Status      string `json:"status" gorm:"size:24;not null;default:open"`
}

func (finding *Finding) BeforeCreate(_ *gorm.DB) error {
	finding.EnsureID()
	return nil
}

type PQCAssessment struct {
	Base
	ScanID                   string   `json:"scan_id" gorm:"type:uuid;not null;uniqueIndex"`
	AssetID                  string   `json:"asset_id" gorm:"type:uuid;not null;index"`
	Asset                    *Asset   `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
	Score                    int      `json:"score"`
	Grade                    string   `json:"grade" gorm:"size:8"`
	QuantumVulnerable        bool     `json:"quantum_vulnerable"`
	RSADependency            bool     `json:"rsa_dependency"`
	ECCDependency            bool     `json:"ecc_dependency"`
	TLS13Enabled             bool     `json:"tls13_enabled"`
	LegacyTLSSupported       bool     `json:"legacy_tls_supported"`
	CertificateRotationReady bool     `json:"certificate_rotation_ready"`
	Rationale                []string `json:"rationale" gorm:"serializer:json"`
}

func (assessment *PQCAssessment) BeforeCreate(_ *gorm.DB) error {
	assessment.EnsureID()
	return nil
}

type AuditLog struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     *string   `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action     string    `json:"action" gorm:"size:80;not null;index"`
	EntityType string    `json:"entity_type" gorm:"size:80;not null;index"`
	EntityID   string    `json:"entity_id,omitempty" gorm:"size:255;index"`
	IPAddress  string    `json:"ip_address,omitempty" gorm:"size:64"`
	UserAgent  string    `json:"user_agent,omitempty" gorm:"type:text"`
	Details    string    `json:"details,omitempty" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`
}

func (auditLog *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if auditLog.ID == "" {
		auditLog.ID = uuid.NewString()
	}
	return nil
}
