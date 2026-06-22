package scoring

import (
	"fmt"
	"strings"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/models"
	"github.com/Prathyusha2909/quantumfield/internal/scanner"
)

var severityWeights = map[string]int{
	"critical": 30,
	"high":     20,
	"medium":   10,
	"low":      5,
	"info":     0,
}

func Analyze(assetID, scanID string, result *scanner.Result, now time.Time) ([]models.Finding, int, models.PQCAssessment) {
	findings := buildFindings(assetID, scanID, result, now)
	riskScore := 0
	for _, finding := range findings {
		riskScore += severityWeights[finding.Severity]
	}
	if riskScore > 100 {
		riskScore = 100
	}

	assessment := assessPQC(assetID, scanID, result, now)
	return findings, riskScore, assessment
}

func buildFindings(assetID, scanID string, result *scanner.Result, now time.Time) []models.Finding {
	certificate := result.Certificate
	findings := make([]models.Finding, 0, 8)
	add := func(findingType, severity, title, description, remediation, evidence string) {
		findings = append(findings, models.Finding{
			AssetID:     assetID,
			ScanID:      scanID,
			Type:        findingType,
			Severity:    severity,
			Title:       title,
			Description: description,
			Remediation: remediation,
			Evidence:    evidence,
			Status:      "open",
		})
	}

	daysRemaining := int(certificate.NotAfter.Sub(now).Hours() / 24)
	switch {
	case certificate.NotAfter.Before(now):
		add("certificate_expired", "critical", "Certificate has expired",
			"The endpoint presents a certificate that is no longer valid.",
			"Replace the certificate immediately and verify automated renewal.",
			fmt.Sprintf("Expired at %s", certificate.NotAfter.UTC().Format(time.RFC3339)))
	case daysRemaining <= 30:
		add("certificate_expiring", "high", "Certificate expires within 30 days",
			"An imminent certificate expiry can cause an availability incident.",
			"Renew the certificate and validate the deployment before the expiry date.",
			fmt.Sprintf("%d days remaining", daysRemaining))
	case daysRemaining <= 90:
		add("certificate_expiring", "medium", "Certificate expires within 90 days",
			"The certificate is approaching its renewal window.",
			"Confirm ownership, renewal automation, and the deployment runbook.",
			fmt.Sprintf("%d days remaining", daysRemaining))
	}

	if !certificate.ChainValid {
		add("invalid_chain", "critical", "Certificate chain is not trusted",
			"The server did not present a chain that validates to a trusted root.",
			"Install the correct intermediate certificates and verify trust from multiple clients.",
			"Issuer: "+certificate.Issuer)
	}
	if !certificate.HostnameValid {
		add("hostname_mismatch", "critical", "Certificate hostname mismatch",
			"The certificate names do not cover the scanned domain.",
			"Issue and deploy a certificate whose SAN extension contains the domain.",
			"Certificate SANs: "+strings.Join(certificate.SubjectAltNames, ", "))
	}
	if result.HSTSChecked && strings.TrimSpace(result.HSTSHeader) == "" {
		add("missing_hsts", "medium", "HTTP Strict Transport Security is missing",
			"The HTTPS response does not instruct browsers to enforce future HTTPS connections.",
			"Add a Strict-Transport-Security header after confirming all subdomains are HTTPS-ready.",
			"No Strict-Transport-Security header observed")
	}
	if result.SupportsTLS10 || result.SupportsTLS11 || result.TLSVersion == "TLS 1.0" || result.TLSVersion == "TLS 1.1" {
		versions := make([]string, 0, 2)
		if result.SupportsTLS10 {
			versions = append(versions, "TLS 1.0")
		}
		if result.SupportsTLS11 {
			versions = append(versions, "TLS 1.1")
		}
		add("legacy_tls", "high", "Legacy TLS protocol is enabled",
			"The endpoint accepts a deprecated TLS version with weaker security properties.",
			"Disable TLS 1.0 and TLS 1.1; require TLS 1.2 or TLS 1.3.",
			"Accepted legacy versions: "+strings.Join(versions, ", "))
	}

	algorithm := strings.ToUpper(certificate.PublicKeyAlgorithm)
	if strings.Contains(algorithm, "RSA") && certificate.KeySize > 0 && certificate.KeySize < 2048 {
		add("weak_key", "high", "RSA public key is too small",
			"The certificate uses an RSA key below the modern 2048-bit baseline.",
			"Reissue the certificate with RSA 2048+ or an approved elliptic curve while planning PQC migration.",
			fmt.Sprintf("RSA key size: %d bits", certificate.KeySize))
	}
	if strings.Contains(algorithm, "ECDSA") && certificate.KeySize > 0 && certificate.KeySize < 256 {
		add("weak_key", "high", "Elliptic-curve key is too small",
			"The certificate uses an elliptic curve below the modern 256-bit baseline.",
			"Reissue the certificate with a stronger approved curve while planning PQC migration.",
			fmt.Sprintf("ECC key size: %d bits", certificate.KeySize))
	}

	if isClassicalPublicKey(algorithm) {
		add("quantum_vulnerable_key", "low", "Certificate depends on quantum-vulnerable public-key cryptography",
			"RSA and elliptic-curve public keys are vulnerable to a sufficiently capable cryptographically relevant quantum computer.",
			"Inventory dependencies, shorten certificate lifetimes, and prepare for standards-based hybrid or post-quantum certificates.",
			fmt.Sprintf("%s %d-bit public key", certificate.PublicKeyAlgorithm, certificate.KeySize))
	}

	return findings
}

func assessPQC(assetID, scanID string, result *scanner.Result, now time.Time) models.PQCAssessment {
	certificate := result.Certificate
	publicKey := strings.ToUpper(certificate.PublicKeyAlgorithm)
	signature := strings.ToUpper(certificate.SignatureAlgorithm)
	rsaDependency := strings.Contains(publicKey, "RSA") || strings.Contains(signature, "RSA")
	eccDependency := strings.Contains(publicKey, "ECDSA") || strings.Contains(publicKey, "ED25519") ||
		strings.Contains(signature, "ECDSA") || strings.Contains(signature, "ED25519")
	classicalSignature := strings.Contains(signature, "RSA") || strings.Contains(signature, "ECDSA") ||
		strings.Contains(signature, "DSA") || strings.Contains(signature, "ED25519")
	legacyTLS := result.SupportsTLS10 || result.SupportsTLS11
	tls13 := result.TLSVersion == "TLS 1.3"
	validity := certificate.NotAfter.Sub(certificate.NotBefore)
	rotationReady := validity <= 398*24*time.Hour && certificate.NotAfter.After(now)

	score := 100
	rationale := make([]string, 0, 5)
	if rsaDependency || eccDependency {
		score -= 35
		rationale = append(rationale, "Certificate authentication relies on RSA or elliptic-curve cryptography.")
	} else {
		rationale = append(rationale, "No recognized RSA/ECC certificate public-key dependency was detected.")
	}
	if classicalSignature {
		score -= 20
		rationale = append(rationale, "The certificate signature uses a classical, non-PQC algorithm.")
	}
	if !tls13 {
		score -= 15
		rationale = append(rationale, "The negotiated connection did not use TLS 1.3.")
	}
	if legacyTLS {
		score -= 15
		rationale = append(rationale, "The endpoint still accepts TLS 1.0 or TLS 1.1.")
	}
	if !rotationReady {
		score -= 15
		rationale = append(rationale, "Certificate lifecycle does not meet the platform's crypto-agility baseline.")
	}
	if score < 0 {
		score = 0
	}

	return models.PQCAssessment{
		AssetID:                  assetID,
		ScanID:                   scanID,
		Score:                    score,
		Grade:                    grade(score),
		QuantumVulnerable:        rsaDependency || eccDependency,
		RSADependency:            rsaDependency,
		ECCDependency:            eccDependency,
		TLS13Enabled:             tls13,
		LegacyTLSSupported:       legacyTLS,
		CertificateRotationReady: rotationReady,
		Rationale:                rationale,
	}
}

func isClassicalPublicKey(algorithm string) bool {
	return strings.Contains(algorithm, "RSA") ||
		strings.Contains(algorithm, "ECDSA") ||
		strings.Contains(algorithm, "DSA") ||
		strings.Contains(algorithm, "ED25519")
}

func grade(score int) string {
	switch {
	case score >= 80:
		return "A"
	case score >= 65:
		return "B"
	case score >= 50:
		return "C"
	case score >= 30:
		return "D"
	default:
		return "F"
	}
}
