package database

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/migrations"
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

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

func migrate(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return err
	}

	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	return db.Transaction(func(transaction *gorm.DB) error {
		// API and worker may start together. The transaction-scoped lock serializes migration runners.
		if err := transaction.Exec("SELECT pg_advisory_xact_lock(?)", int64(717551234)).Error; err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				continue
			}
			var count int64
			if err := transaction.Table("schema_migrations").Where("version = ?", entry.Name()).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			sql, err := migrations.Files.ReadFile(entry.Name())
			if err != nil {
				return err
			}
			if err := transaction.Exec(string(sql)).Error; err != nil {
				return fmt.Errorf("%s: %w", entry.Name(), err)
			}
			if err := transaction.Exec("INSERT INTO schema_migrations (version) VALUES (?)", entry.Name()).Error; err != nil {
				return err
			}
			log.Printf("applied database migration %s", entry.Name())
		}
		return nil
	})
}

func SeedDemo(db *gorm.DB) {
	const email = "demo@quantumfield.dev"
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte("QuantumField123!"), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Printf("demo seed skipped: %v", hashErr)
			return
		}
		user = models.User{
			Name:         "Demo Analyst",
			Email:        email,
			PasswordHash: string(passwordHash),
			Role:         models.RoleAnalyst,
		}
		if createErr := db.Create(&user).Error; createErr != nil && !strings.Contains(strings.ToLower(createErr.Error()), "duplicate") {
			log.Printf("demo seed skipped: %v", createErr)
			return
		} else if createErr != nil {
			if loadErr := db.Where("email = ?", email).First(&user).Error; loadErr != nil {
				log.Printf("demo seed skipped after concurrent insert: %v", loadErr)
				return
			}
		}
	}

	demoAssets := []models.Asset{
		{UserID: user.ID, Domain: "example.com", Port: 443, Label: "Safe baseline", Status: "pending"},
		{UserID: user.ID, Domain: "expired.badssl.com", Port: 443, Label: "Expired certificate test", Status: "pending"},
		{UserID: user.ID, Domain: "wrong.host.badssl.com", Port: 443, Label: "Hostname mismatch test", Status: "pending"},
	}
	for _, asset := range demoAssets {
		var count int64
		db.Model(&models.Asset{}).
			Where("user_id = ? AND domain = ? AND port = ?", user.ID, asset.Domain, asset.Port).
			Count(&count)
		if count == 0 {
			if err := db.Create(&asset).Error; err != nil {
				log.Printf("could not seed demo asset %s: %v", asset.Domain, err)
			}
		}
	}
}
