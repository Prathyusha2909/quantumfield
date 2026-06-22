package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/internal/queue"
	"github.com/Prathyusha2909/quantumfield/internal/scanner"
	"github.com/Prathyusha2909/quantumfield/internal/scoring"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Worker struct {
	DB      *gorm.DB
	Queue   *queue.Client
	Scanner *scanner.Scanner
}

func (worker *Worker) Run(ctx context.Context) error {
	log.Print("scan worker is ready")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		job, err := worker.Queue.Dequeue(ctx, 5*time.Second)
		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
				continue
			}
			log.Printf("dequeue failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := worker.process(ctx, *job); err != nil {
			log.Printf("scan %s failed: %v", job.ScanID, err)
		}
	}
}

func (worker *Worker) process(parent context.Context, job queue.ScanJob) error {
	var scan models.Scan
	if err := worker.DB.First(&scan, "id = ? AND asset_id = ?", job.ScanID, job.AssetID).Error; err != nil {
		return fmt.Errorf("load scan: %w", err)
	}
	var asset models.Asset
	if err := worker.DB.First(&asset, "id = ? AND user_id = ?", job.AssetID, job.UserID).Error; err != nil {
		worker.fail(job, "asset no longer exists")
		return fmt.Errorf("load asset: %w", err)
	}

	started := time.Now().UTC()
	if err := worker.DB.Model(&scan).Updates(map[string]any{
		"status":        models.ScanRunning,
		"started_at":    &started,
		"error_message": "",
	}).Error; err != nil {
		return fmt.Errorf("mark scan running: %w", err)
	}
	worker.DB.Model(&asset).Update("status", models.ScanRunning)

	result, err := worker.Scanner.Scan(parent, asset.Domain, asset.Port)
	if err != nil {
		worker.fail(job, err.Error())
		return err
	}

	completed := time.Now().UTC()
	findings, riskScore, assessment := scoring.Analyze(asset.ID, scan.ID, result, completed)
	certificate := result.Certificate
	certificate.AssetID = asset.ID
	certificate.ScanID = scan.ID
	assessment.AssetID = asset.ID
	assessment.ScanID = scan.ID

	err = worker.DB.Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Create(&certificate).Error; err != nil {
			return err
		}
		if len(findings) > 0 {
			if err := transaction.Create(&findings).Error; err != nil {
				return err
			}
		}
		if err := transaction.Create(&assessment).Error; err != nil {
			return err
		}
		if err := transaction.Model(&scan).Updates(map[string]any{
			"status":       models.ScanCompleted,
			"completed_at": &completed,
			"duration_ms":  completed.Sub(started).Milliseconds(),
			"risk_score":   riskScore,
			"pqc_score":    assessment.Score,
			"tls_version":  result.TLSVersion,
			"cipher_suite": result.CipherSuite,
		}).Error; err != nil {
			return err
		}
		return transaction.Model(&asset).Updates(map[string]any{
			"status":             models.ScanCompleted,
			"last_scanned_at":    &completed,
			"current_risk_score": riskScore,
			"current_pqc_score":  assessment.Score,
		}).Error
	})
	if err != nil {
		worker.fail(job, "could not persist scan results")
		return fmt.Errorf("persist scan: %w", err)
	}

	log.Printf("scan %s completed for %s:%d (risk=%d pqc=%d)", scan.ID, asset.Domain, asset.Port, riskScore, assessment.Score)
	return nil
}

func (worker *Worker) fail(job queue.ScanJob, message string) {
	completed := time.Now().UTC()
	worker.DB.Model(&models.Scan{}).Where("id = ?", job.ScanID).Updates(map[string]any{
		"status":        models.ScanFailed,
		"completed_at":  &completed,
		"error_message": message,
	})
	worker.DB.Model(&models.Asset{}).Where("id = ?", job.AssetID).Update("status", models.ScanFailed)
}
