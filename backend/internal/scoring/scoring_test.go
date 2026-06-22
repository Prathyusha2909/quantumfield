package scoring

import (
	"testing"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/internal/scanner"
)

func TestAnalyzeFlagsClassicalCertificateAndMissingHSTS(t *testing.T) {
	now := time.Date(2026, time.June, 22, 12, 0, 0, 0, time.UTC)
	result := &scanner.Result{
		Certificate: models.Certificate{
			NotBefore:          now.Add(-30 * 24 * time.Hour),
			NotAfter:           now.Add(60 * 24 * time.Hour),
			PublicKeyAlgorithm: "RSA",
			SignatureAlgorithm: "SHA256-RSA",
			KeySize:            2048,
			ChainValid:         true,
			HostnameValid:      true,
		},
		TLSVersion:  "TLS 1.2",
		HSTSChecked: true,
	}

	findings, risk, assessment := Analyze("asset", "scan", result, now)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	if risk != 25 {
		t.Fatalf("expected risk score 25, got %d", risk)
	}
	if assessment.Score != 30 {
		t.Fatalf("expected PQC score 30, got %d", assessment.Score)
	}
	if !assessment.RSADependency || assessment.Grade != "D" {
		t.Fatalf("unexpected assessment: %+v", assessment)
	}
}

func TestAnalyzeCapsRiskAtOneHundred(t *testing.T) {
	now := time.Now().UTC()
	result := &scanner.Result{
		Certificate: models.Certificate{
			NotBefore:          now.Add(-800 * 24 * time.Hour),
			NotAfter:           now.Add(-400 * 24 * time.Hour),
			PublicKeyAlgorithm: "RSA",
			SignatureAlgorithm: "SHA1-RSA",
			KeySize:            1024,
			ChainValid:         false,
			HostnameValid:      false,
		},
		TLSVersion:    "TLS 1.0",
		HSTSChecked:   true,
		SupportsTLS10: true,
	}

	_, risk, _ := Analyze("asset", "scan", result, now)
	if risk != 100 {
		t.Fatalf("expected capped risk score 100, got %d", risk)
	}
}
