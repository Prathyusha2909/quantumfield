package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/audit"
	"github.com/Prathyusha2909/quantumfield/internal/auth"
	"github.com/Prathyusha2909/quantumfield/internal/middleware"
	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/internal/queue"
	"github.com/Prathyusha2909/quantumfield/internal/target"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB             *gorm.DB
	Queue          *queue.Client
	Auth           *auth.Service
	MaxScanRetries int
}

type credentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

func (handler *Handler) Health(context *gin.Context) {
	sqlDB, err := handler.DB.DB()
	if err != nil || sqlDB.PingContext(context.Request.Context()) != nil {
		context.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "unavailable"})
		return
	}
	if err := handler.Queue.Ping(context.Request.Context()); err != nil {
		context.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "redis": "unavailable"})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "quantumfield-api",
		"time":    time.Now().UTC(),
	})
}

func (handler *Handler) Register(context *gin.Context) {
	var request struct {
		Name string `json:"name" binding:"required,min=2,max=120"`
		credentials
	}
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "provide a valid name, email, and password of at least 8 characters"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	var count int64
	handler.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		context.JSON(http.StatusConflict, gin.H{"error": "an account with this email already exists"})
		return
	}

	hash, err := handler.Auth.HashPassword(request.Password)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not secure password"})
		return
	}
	user := models.User{
		Name:         strings.TrimSpace(request.Name),
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleAnalyst,
	}
	if err := handler.DB.Create(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create account"})
		return
	}

	token, err := handler.Auth.CreateToken(user.ID, user.Role)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}
	handler.recordAudit(context, &user.ID, models.AuditUserRegistered, "user", user.ID, "account registered")
	context.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func (handler *Handler) Login(context *gin.Context) {
	var request credentials
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "provide a valid email and password"})
		return
	}

	var user models.User
	if err := handler.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(request.Email))).First(&user).Error; err != nil {
		handler.recordAudit(context, nil, models.AuditLoginFailed, "session", "", "unknown account")
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err := handler.Auth.VerifyPassword(user.PasswordHash, request.Password); err != nil {
		handler.recordAudit(context, &user.ID, models.AuditLoginFailed, "session", user.ID, "invalid password")
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := handler.Auth.CreateToken(user.ID, user.Role)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}
	handler.recordAudit(context, &user.ID, models.AuditLoginSuccess, "session", user.ID, "login succeeded")
	context.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (handler *Handler) Me(context *gin.Context) {
	var user models.User
	if err := handler.DB.First(&user, "id = ?", userID(context)).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (handler *Handler) CreateAsset(context *gin.Context) {
	var request struct {
		Domain string `json:"domain" binding:"required"`
		Port   int    `json:"port"`
		Label  string `json:"label" binding:"max=120"`
	}
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
		return
	}
	domain, port, err := target.Normalize(request.Domain, request.Port)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset := models.Asset{
		UserID: userID(context),
		Domain: domain,
		Port:   port,
		Label:  strings.TrimSpace(request.Label),
		Status: "pending",
	}
	if err := handler.DB.Create(&asset).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			context.JSON(http.StatusConflict, gin.H{"error": "this domain and port are already in your inventory"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create asset"})
		return
	}
	handler.recordAudit(context, stringPointer(asset.UserID), models.AuditAssetCreated, "asset", asset.ID, asset.Domain)
	context.JSON(http.StatusCreated, asset)
}

func (handler *Handler) ListAssets(context *gin.Context) {
	var assets []models.Asset
	handler.DB.Where("user_id = ?", userID(context)).Order("created_at DESC").Find(&assets)
	context.JSON(http.StatusOK, assets)
}

func (handler *Handler) GetAsset(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	handler.DB.Where("asset_id = ?", asset.ID).
		Order("created_at DESC").
		Limit(20).
		Preload("Certificate").
		Preload("PQCAssessment").
		Find(&asset.Scans)
	context.JSON(http.StatusOK, asset)
}

func (handler *Handler) DeleteAsset(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	if err := handler.DB.Delete(&asset).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete asset"})
		return
	}
	context.Status(http.StatusNoContent)
}

func (handler *Handler) StartScan(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	var active int64
	handler.DB.Model(&models.Scan{}).
		Where("asset_id = ? AND status IN ?", asset.ID, []string{models.ScanQueued, models.ScanRunning}).
		Count(&active)
	if active > 0 {
		context.JSON(http.StatusConflict, gin.H{"error": "a scan is already queued or running for this asset"})
		return
	}

	maxRetries := handler.MaxScanRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	scan := models.Scan{AssetID: asset.ID, Status: models.ScanQueued, MaxRetries: maxRetries}
	if err := handler.DB.Create(&scan).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "could not create scan"})
		return
	}

	job := queue.ScanJob{ScanID: scan.ID, AssetID: asset.ID, UserID: asset.UserID, Attempt: 0}
	if err := handler.Queue.Enqueue(context.Request.Context(), job); err != nil {
		failedAt := time.Now().UTC()
		handler.DB.Model(&scan).Updates(map[string]any{
			"status":        models.ScanFailed,
			"error_message": "failed to enqueue scan",
			"last_error":    "failed to enqueue scan",
			"failed_at":     &failedAt,
			"completed_at":  &failedAt,
		})
		handler.recordAudit(context, stringPointer(asset.UserID), models.AuditScanFailed, "scan", scan.ID, "failed to enqueue scan")
		context.JSON(http.StatusServiceUnavailable, gin.H{"error": "scan queue is unavailable"})
		return
	}
	handler.DB.Model(&asset).Update("status", models.ScanQueued)
	handler.recordAudit(context, stringPointer(asset.UserID), models.AuditScanQueued, "scan", scan.ID, asset.Domain)
	context.JSON(http.StatusAccepted, scan)
}

func (handler *Handler) ListScans(context *gin.Context) {
	var scans []models.Scan
	handler.DB.Joins("JOIN assets ON assets.id = scans.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", userID(context)).
		Preload("Asset").
		Preload("Certificate").
		Preload("PQCAssessment").
		Order("scans.created_at DESC").
		Limit(200).
		Find(&scans)
	context.JSON(http.StatusOK, scans)
}

func (handler *Handler) GetScan(context *gin.Context) {
	var scan models.Scan
	err := handler.DB.Joins("JOIN assets ON assets.id = scans.asset_id").
		Where("scans.id = ? AND assets.user_id = ? AND assets.deleted_at IS NULL", context.Param("id"), userID(context)).
		Preload("Asset").
		Preload("Certificate").
		Preload("Findings", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Preload("PQCAssessment").
		First(&scan).Error
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}
	context.JSON(http.StatusOK, scan)
}

func (handler *Handler) GetLatestCertificate(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	var certificate models.Certificate
	err = handler.DB.Where("asset_id = ?", asset.ID).Order("created_at DESC").First(&certificate).Error
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "no certificate scan is available"})
		return
	}
	context.JSON(http.StatusOK, certificate)
}

func (handler *Handler) GetAssetFindings(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	var findings []models.Finding
	handler.DB.Where("asset_id = ?", asset.ID).Order("created_at DESC").Find(&findings)
	context.JSON(http.StatusOK, findings)
}

func (handler *Handler) GetAssetPQC(context *gin.Context) {
	asset, err := handler.findAsset(context)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	var assessment models.PQCAssessment
	err = handler.DB.Where("asset_id = ?", asset.ID).Order("created_at DESC").First(&assessment).Error
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "no PQC assessment is available"})
		return
	}
	context.JSON(http.StatusOK, assessment)
}

func (handler *Handler) ListFindings(context *gin.Context) {
	var findings []models.Finding
	query := handler.DB.Joins("JOIN assets ON assets.id = findings.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", userID(context)).
		Preload("Asset").
		Order("CASE findings.severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END, findings.created_at DESC")
	if severity := strings.TrimSpace(context.Query("severity")); severity != "" {
		query = query.Where("findings.severity = ?", severity)
	}
	query.Limit(500).Find(&findings)
	context.JSON(http.StatusOK, findings)
}

func (handler *Handler) ListCertificates(context *gin.Context) {
	var certificates []models.Certificate
	handler.DB.Joins("JOIN assets ON assets.id = certificates.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", userID(context)).
		Preload("Asset").
		Order("certificates.created_at DESC").
		Limit(500).
		Find(&certificates)
	context.JSON(http.StatusOK, certificates)
}

func (handler *Handler) ListPQCAssessments(context *gin.Context) {
	var assessments []models.PQCAssessment
	handler.DB.Joins("JOIN assets ON assets.id = pqc_assessments.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", userID(context)).
		Preload("Asset").
		Order("pqc_assessments.created_at DESC").
		Limit(500).
		Find(&assessments)
	context.JSON(http.StatusOK, assessments)
}

func (handler *Handler) Dashboard(context *gin.Context) {
	uid := userID(context)
	var assets []models.Asset
	handler.DB.Where("user_id = ?", uid).Order("current_risk_score DESC").Find(&assets)

	var recentScans []models.Scan
	handler.DB.Joins("JOIN assets ON assets.id = scans.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", uid).
		Preload("Asset").
		Order("scans.created_at DESC").
		Limit(8).
		Find(&recentScans)

	var findings []models.Finding
	handler.DB.Joins("JOIN assets ON assets.id = findings.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL AND findings.status = ?", uid, "open").
		Order("CASE findings.severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, findings.created_at DESC").
		Limit(8).
		Find(&findings)

	totalRisk := 0
	totalPQC := 0
	assessed := 0
	critical := 0
	for _, asset := range assets {
		totalRisk += asset.CurrentRiskScore
		if asset.LastScannedAt != nil {
			totalPQC += asset.CurrentPQCScore
			assessed++
		}
	}
	for _, finding := range findings {
		if finding.Severity == "critical" {
			critical++
		}
	}
	averageRisk, averagePQC := 0, 0
	if assessed > 0 {
		averageRisk = totalRisk / assessed
		averagePQC = totalPQC / assessed
	}

	context.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"asset_count":        len(assets),
			"assessed_count":     assessed,
			"average_risk_score": averageRisk,
			"average_pqc_score":  averagePQC,
			"critical_findings":  critical,
		},
		"assets":            assets,
		"recent_scans":      recentScans,
		"priority_findings": findings,
	})
}

func (handler *Handler) ReportSummary(context *gin.Context) {
	context.JSON(http.StatusOK, handler.buildReport(userID(context)))
}

func (handler *Handler) ExportReport(context *gin.Context) {
	uid := userID(context)
	report := handler.buildReport(uid)
	handler.recordAudit(context, stringPointer(uid), models.AuditReportExported, "report", "", "JSON portfolio report")
	context.Header("Content-Disposition", "attachment; filename=quantumfield-report.json")
	context.JSON(http.StatusOK, report)
}

type reportSummary struct {
	GeneratedAt                time.Time      `json:"generated_at"`
	FindingsBySeverity         []severityRow  `json:"findings_by_severity"`
	CertificatesByAlgorithm    []algorithmRow `json:"certificates_by_algorithm"`
	CertificatesExpiring90Days int64          `json:"certificates_expiring_90_days"`
}

type severityRow struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

type algorithmRow struct {
	Algorithm string `json:"algorithm"`
	Count     int64  `json:"count"`
}

func (handler *Handler) buildReport(uid string) reportSummary {
	var severities []severityRow
	handler.DB.Model(&models.Finding{}).
		Select("severity, count(*) as count").
		Joins("JOIN assets ON assets.id = findings.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", uid).
		Group("severity").
		Scan(&severities)

	var algorithms []algorithmRow
	handler.DB.Model(&models.Certificate{}).
		Select("public_key_algorithm as algorithm, count(*) as count").
		Joins("JOIN assets ON assets.id = certificates.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL", uid).
		Group("public_key_algorithm").
		Scan(&algorithms)

	var expiring int64
	handler.DB.Model(&models.Certificate{}).
		Joins("JOIN assets ON assets.id = certificates.asset_id").
		Where("assets.user_id = ? AND assets.deleted_at IS NULL AND certificates.not_after BETWEEN ? AND ?",
			uid, time.Now(), time.Now().Add(90*24*time.Hour)).
		Count(&expiring)

	return reportSummary{
		GeneratedAt:                time.Now().UTC(),
		FindingsBySeverity:         severities,
		CertificatesByAlgorithm:    algorithms,
		CertificatesExpiring90Days: expiring,
	}
}

func (handler *Handler) ListAuditLogs(context *gin.Context) {
	var logs []models.AuditLog
	handler.DB.
		Where("user_id = ?", userID(context)).
		Order("created_at DESC").
		Limit(200).
		Find(&logs)
	context.JSON(http.StatusOK, logs)
}

func (handler *Handler) findAsset(context *gin.Context) (models.Asset, error) {
	var asset models.Asset
	err := handler.DB.Where("id = ? AND user_id = ?", context.Param("id"), userID(context)).First(&asset).Error
	return asset, err
}

func userID(context *gin.Context) string {
	return context.GetString(middleware.ContextUserID)
}

func (handler *Handler) recordAudit(
	context *gin.Context,
	uid *string,
	action,
	entityType,
	entityID,
	details string,
) {
	audit.Record(handler.DB, audit.Event{
		UserID:     uid,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  context.ClientIP(),
		UserAgent:  context.Request.UserAgent(),
		Details:    details,
	})
}

func stringPointer(value string) *string {
	return &value
}
