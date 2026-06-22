package httpapi

import (
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/config"
	"github.com/Prathyusha2909/quantumfield/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *Handler, cfg config.Config) *gin.Engine {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		panic(err)
	}
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{
			"Content-Length",
			"Retry-After",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", handler.Health)
	api := router.Group("/api")
	api.GET("/health", handler.Health)

	authRoutes := api.Group("/auth")
	authRateLimit := middleware.RateLimit(handler.Queue, "auth", 10, time.Minute, middleware.ClientIP)
	authRoutes.POST("/register", authRateLimit, handler.Register)
	authRoutes.POST("/login", authRateLimit, handler.Login)

	protected := api.Group("")
	protected.Use(middleware.Authenticate(handler.Auth))
	protected.GET("/auth/me", handler.Me)
	protected.GET("/dashboard", handler.Dashboard)
	protected.GET("/reports/summary", handler.ReportSummary)
	protected.GET("/reports/export", handler.ExportReport)
	protected.GET("/audit-logs", handler.ListAuditLogs)

	protected.POST("/assets", handler.CreateAsset)
	protected.GET("/assets", handler.ListAssets)
	protected.GET("/assets/:id", handler.GetAsset)
	protected.DELETE("/assets/:id", handler.DeleteAsset)
	protected.POST(
		"/assets/:id/scan",
		middleware.RateLimit(handler.Queue, "scan", 10, 10*time.Minute, middleware.AuthenticatedUser),
		handler.StartScan,
	)
	protected.GET("/assets/:id/certificate", handler.GetLatestCertificate)
	protected.GET("/assets/:id/findings", handler.GetAssetFindings)
	protected.GET("/assets/:id/pqc-assessment", handler.GetAssetPQC)

	protected.GET("/scans", handler.ListScans)
	protected.GET("/scans/:id", handler.GetScan)
	protected.GET("/findings", handler.ListFindings)
	protected.GET("/certificates", handler.ListCertificates)
	protected.GET("/pqc-assessments", handler.ListPQCAssessments)

	return router
}
