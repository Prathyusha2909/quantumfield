package scanner

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/models"
)

type Result struct {
	Certificate   models.Certificate
	TLSVersion    string
	CipherSuite   string
	HSTSHeader    string
	HSTSChecked   bool
	SupportsTLS10 bool
	SupportsTLS11 bool
}

type Scanner struct {
	timeout time.Duration
}

func New(timeout time.Duration) *Scanner {
	return &Scanner{timeout: timeout}
}

func (scanner *Scanner) Scan(ctx context.Context, domain string, port int) (*Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if err := validateTarget(ctx, domain); err != nil {
		return nil, err
	}

	address := net.JoinHostPort(domain, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: scanner.timeout}
	connection, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		ServerName:         domain,
		InsecureSkipVerify: true, // Verification is performed explicitly below so invalid chains can still be inventoried.
		MinVersion:         tls.VersionTLS10,
	})
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}
	defer connection.Close()

	state := connection.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("server did not provide a certificate")
	}

	leaf := state.PeerCertificates[0]
	intermediates := x509.NewCertPool()
	for _, certificate := range state.PeerCertificates[1:] {
		intermediates.AddCert(certificate)
	}

	_, chainErr := leaf.Verify(x509.VerifyOptions{Intermediates: intermediates})
	hostnameErr := leaf.VerifyHostname(domain)

	fingerprint := sha256.Sum256(leaf.Raw)
	certificate := models.Certificate{
		Subject:            leaf.Subject.String(),
		CommonName:         leaf.Subject.CommonName,
		Issuer:             leaf.Issuer.String(),
		SerialNumber:       strings.ToUpper(leaf.SerialNumber.Text(16)),
		NotBefore:          leaf.NotBefore,
		NotAfter:           leaf.NotAfter,
		PublicKeyAlgorithm: leaf.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: leaf.SignatureAlgorithm.String(),
		KeySize:            publicKeySize(leaf),
		ChainValid:         chainErr == nil,
		HostnameValid:      hostnameErr == nil,
		SelfSigned:         bytes.Equal(leaf.RawIssuer, leaf.RawSubject) && leaf.CheckSignatureFrom(leaf) == nil,
		SubjectAltNames:    leaf.DNSNames,
		FingerprintSHA256:  colonHex(fingerprint[:]),
		PEM:                string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw})),
	}

	hstsHeader, hstsChecked := scanner.checkHSTS(domain, port)

	return &Result{
		Certificate:   certificate,
		TLSVersion:    tlsVersionName(state.Version),
		CipherSuite:   tls.CipherSuiteName(state.CipherSuite),
		HSTSHeader:    hstsHeader,
		HSTSChecked:   hstsChecked,
		SupportsTLS10: scanner.supportsVersion(address, domain, tls.VersionTLS10),
		SupportsTLS11: scanner.supportsVersion(address, domain, tls.VersionTLS11),
	}, nil
}

func validateTarget(ctx context.Context, domain string) error {
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("target did not resolve to an IP address")
	}
	for _, address := range addresses {
		ip := address.IP
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsMulticast() {
			return fmt.Errorf("target resolves to a private or reserved network address")
		}
	}
	return nil
}

func (scanner *Scanner) checkHSTS(domain string, port int) (string, bool) {
	host := domain
	if port != 443 {
		host = net.JoinHostPort(domain, strconv.Itoa(port))
	}
	client := &http.Client{
		Timeout: scanner.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:         domain,
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	request, err := http.NewRequest(http.MethodGet, "https://"+host+"/", nil)
	if err != nil {
		return "", false
	}
	request.Header.Set("User-Agent", "QuantumField-TLS-Scanner/1.0")
	response, err := client.Do(request)
	if err != nil {
		return "", false
	}
	defer response.Body.Close()
	_, _ = io.CopyN(io.Discard, response.Body, 1024)
	return response.Header.Get("Strict-Transport-Security"), true
}

func (scanner *Scanner) supportsVersion(address, domain string, version uint16) bool {
	connection, err := tls.DialWithDialer(
		&net.Dialer{Timeout: scanner.timeout / 2},
		"tcp",
		address,
		&tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: true,
			MinVersion:         version,
			MaxVersion:         version,
		},
	)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func publicKeySize(certificate *x509.Certificate) int {
	switch key := certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		return key.N.BitLen()
	case *ecdsa.PublicKey:
		return key.Curve.Params().BitSize
	case ed25519.PublicKey:
		return len(key) * 8
	default:
		return 0
	}
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return fmt.Sprintf("Unknown (0x%x)", version)
	}
}

func colonHex(value []byte) string {
	encoded := strings.ToUpper(hex.EncodeToString(value))
	parts := make([]string, 0, len(encoded)/2)
	for index := 0; index < len(encoded); index += 2 {
		parts = append(parts, encoded[index:index+2])
	}
	return strings.Join(parts, ":")
}
