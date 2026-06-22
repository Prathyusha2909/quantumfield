package audit

import (
	"log"
	"strings"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"gorm.io/gorm"
)

type Event struct {
	UserID     *string
	Action     string
	EntityType string
	EntityID   string
	IPAddress  string
	UserAgent  string
	Details    string
}

func Record(db *gorm.DB, event Event) {
	entry := models.AuditLog{
		UserID:     event.UserID,
		Action:     event.Action,
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		IPAddress:  truncate(event.IPAddress, 64),
		UserAgent:  truncate(event.UserAgent, 1024),
		Details:    truncate(event.Details, 2048),
	}
	if err := db.Create(&entry).Error; err != nil {
		log.Printf("audit event %s could not be persisted: %v", event.Action, err)
	}
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
