//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/auth"
	"github.com/Prathyusha2909/quantumfield/internal/config"
	"github.com/Prathyusha2909/quantumfield/internal/database"
	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/internal/queue"
)

func TestRegisterCreateAssetAndQueueScan(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	redisAddress := os.Getenv("TEST_REDIS_ADDR")
	if databaseURL == "" || redisAddress == "" {
		t.Skip("TEST_DATABASE_URL and TEST_REDIS_ADDR are required")
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	queueClient := queue.New(redisAddress, "", 15)
	defer queueClient.Close()

	authService := auth.New("integration-secret-that-is-long-enough", time.Hour)
	router := NewRouter(&Handler{DB: db, Queue: queueClient, Auth: authService}, config.Config{
		CORSOrigins: []string{"http://localhost"},
	})

	email := fmt.Sprintf("integration-%d@quantumfield.dev", time.Now().UnixNano())
	registerResponse := performJSONRequest(t, router, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Integration Analyst",
		"email":    email,
		"password": "Integration123!",
	})
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register returned %d: %s", registerResponse.Code, registerResponse.Body.String())
	}
	var session struct {
		Token string      `json:"token"`
		User  models.User `json:"user"`
	}
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Token == "" || session.User.ID == "" {
		t.Fatalf("register response did not include a session: %+v", session)
	}

	assetResponse := performJSONRequest(t, router, http.MethodPost, "/api/assets", session.Token, map[string]any{
		"domain": "https://example.com/path",
		"port":   443,
		"label":  "Integration target",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("create asset returned %d: %s", assetResponse.Code, assetResponse.Body.String())
	}
	var asset models.Asset
	if err := json.Unmarshal(assetResponse.Body.Bytes(), &asset); err != nil {
		t.Fatal(err)
	}
	if asset.Domain != "example.com" || asset.UserID != session.User.ID {
		t.Fatalf("unexpected asset: %+v", asset)
	}

	scanResponse := performJSONRequest(t, router, http.MethodPost, "/api/assets/"+asset.ID+"/scan", session.Token, nil)
	if scanResponse.Code != http.StatusAccepted {
		t.Fatalf("queue scan returned %d: %s", scanResponse.Code, scanResponse.Body.String())
	}
	var scan models.Scan
	if err := json.Unmarshal(scanResponse.Body.Bytes(), &scan); err != nil {
		t.Fatal(err)
	}
	if scan.Status != models.ScanQueued {
		t.Fatalf("expected queued scan, got %+v", scan)
	}

	dequeueContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	job, err := queueClient.Dequeue(dequeueContext, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if job.ScanID != scan.ID || job.AssetID != asset.ID || job.UserID != session.User.ID {
		t.Fatalf("unexpected queued job: %+v", job)
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, &payload)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
