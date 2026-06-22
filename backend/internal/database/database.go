package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, fmt.Errorf("enable pgcrypto: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Asset{},
		&models.Scan{},
		&models.Certificate{},
		&models.Finding{},
		&models.PQCAssessment{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

func SeedDemo(db *gorm.DB) {
	const email = "demo@quantumfield.dev"
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("QuantumField123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("demo seed skipped: %v", err)
		return
	}

	user := models.User{
		Name:         "Demo Analyst",
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         models.RoleAnalyst,
	}
	if err := db.Create(&user).Error; err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		log.Printf("demo seed skipped: %v", err)
	}
}
