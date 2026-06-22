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
	"net/netip"
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
	timeout  time.Duration
	resolver resolver
}

func New(timeout time.Duration) *Scanner {
	return &Scanner{
		timeout:  timeout,
		resolver: net.DefaultResolver,
	}
}

func (scanner *Scanner) Scan(ctx context.Context, domain string, port int) (*Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	target, err := scanner.resolveTarget(ctx, domain, port)
	if err != nil {
		return nil, err
	}

	connection, err := scanner.dialTLS(ctx, target, tls.VersionTLS10, 0)
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

	hstsHeader, hstsChecked := scanner.checkHSTS(ctx, target)

	return &Result{
		Certificate:   certificate,
		TLSVersion:    tlsVersionName(state.Version),
		CipherSuite:   tls.CipherSuiteName(state.CipherSuite),
		HSTSHeader:    hstsHeader,
		HSTSChecked:   hstsChecked,
		SupportsTLS10: scanner.supportsVersion(ctx, target, tls.VersionTLS10),
		SupportsTLS11: scanner.supportsVersion(ctx, target, tls.VersionTLS11),
	}, nil
}

type resolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type resolvedTarget struct {
	domain  string
	port    int
	ip      net.IP
	address string
}

var blockedNetworkPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001:db8::/32"),
}

func (scanner *Scanner) resolveTarget(ctx context.Context, domain string, port int) (resolvedTarget, error) {
	addresses, err := scanner.resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return resolvedTarget{}, fmt.Errorf("resolve target: %w", err)
	}
	if len(addresses) == 0 {
		return resolvedTarget{}, fmt.Errorf("target did not resolve to an IP address")
	}

	for _, address := range addresses {
		if !isPublicIP(address.IP) {
			return resolvedTarget{}, fmt.Errorf("target resolves to a private or reserved network address")
		}
	}

	pinnedIP := append(net.IP(nil), addresses[0].IP...)
	return resolvedTarget{
		domain:  domain,
		port:    port,
		ip:      pinnedIP,
		address: net.JoinHostPort(pinnedIP.String(), strconv.Itoa(port)),
	}, nil
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}

	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	for _, prefix := range blockedNetworkPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func (scanner *Scanner) dialTLS(ctx context.Context, target resolvedTarget, minVersion, maxVersion uint16) (*tls.Conn, error) {
	dialContext, cancel := context.WithTimeout(ctx, scanner.timeout)
	defer cancel()

	rawConnection, err := (&net.Dialer{Timeout: scanner.timeout}).DialContext(dialContext, "tcp", target.address)
	if err != nil {
		return nil, err
	}

	connection := tls.Client(rawConnection, &tls.Config{
		ServerName:         target.domain,
		InsecureSkipVerify: true, // Verification is performed explicitly so invalid certificates can still be inventoried.
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
	})
	if err := connection.HandshakeContext(dialContext); err != nil {
		_ = rawConnection.Close()
		return nil, err
	}
	return connection, nil
}

func (scanner *Scanner) checkHSTS(ctx context.Context, target resolvedTarget) (string, bool) {
	host := target.domain
	if target.port != 443 {
		host = net.JoinHostPort(target.domain, strconv.Itoa(target.port))
	}

	client := &http.Client{
		Timeout: scanner.timeout,
		Transport: &http.Transport{
			DialContext: func(dialContext context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{Timeout: scanner.timeout}).DialContext(dialContext, network, target.address)
			},
			TLSClientConfig: &tls.Config{
				ServerName:         target.domain,
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
			TLSHandshakeTimeout: scanner.timeout,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	request, err := http.NewRequest(http.MethodGet, "https://"+host+"/", nil)
	if err != nil {
		return "", false
	}
	request = request.WithContext(ctx)
	request.Header.Set("User-Agent", "QuantumField-TLS-Scanner/1.0")
	response, err := client.Do(request)
	if err != nil {
		return "", false
	}
	defer response.Body.Close()
	_, _ = io.CopyN(io.Discard, response.Body, 1024)
	return response.Header.Get("Strict-Transport-Security"), true
}

func (scanner *Scanner) supportsVersion(ctx context.Context, target resolvedTarget, version uint16) bool {
	probeContext, cancel := context.WithTimeout(ctx, scanner.timeout/2)
	defer cancel()

	connection, err := scanner.dialTLS(probeContext, target, version, version)
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
