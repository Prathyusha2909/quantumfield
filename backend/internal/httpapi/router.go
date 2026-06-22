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
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", handler.Health)
	api := router.Group("/api")
	api.GET("/health", handler.Health)

	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", handler.Register)
	authRoutes.POST("/login", handler.Login)

	protected := api.Group("")
	protected.Use(middleware.Authenticate(handler.Auth))
	protected.GET("/auth/me", handler.Me)
	protected.GET("/dashboard", handler.Dashboard)
	protected.GET("/reports/summary", handler.ReportSummary)

	protected.POST("/assets", handler.CreateAsset)
	protected.GET("/assets", handler.ListAssets)
	protected.GET("/assets/:id", handler.GetAsset)
	protected.DELETE("/assets/:id", handler.DeleteAsset)
	protected.POST("/assets/:id/scan", handler.StartScan)
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
