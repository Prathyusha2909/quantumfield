package scanner

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type staticResolver struct {
	addresses []net.IPAddr
	err       error
}

func (resolver staticResolver) LookupIPAddr(_ context.Context, _ string) ([]net.IPAddr, error) {
	return resolver.addresses, resolver.err
}

func TestResolveTargetPinsValidatedPublicIP(t *testing.T) {
	tlsScanner := &Scanner{
		timeout: time.Second,
		resolver: staticResolver{addresses: []net.IPAddr{
			{IP: net.ParseIP("1.1.1.1")},
			{IP: net.ParseIP("8.8.8.8")},
		}},
	}

	target, err := tlsScanner.resolveTarget(context.Background(), "example.com", 443)
	if err != nil {
		t.Fatal(err)
	}
	if target.domain != "example.com" {
		t.Fatalf("expected SNI domain example.com, got %s", target.domain)
	}
	if target.address != "1.1.1.1:443" {
		t.Fatalf("expected pinned address 1.1.1.1:443, got %s", target.address)
	}
}

func TestResolveTargetRejectsMixedPublicAndPrivateAnswers(t *testing.T) {
	tlsScanner := &Scanner{
		timeout: time.Second,
		resolver: staticResolver{addresses: []net.IPAddr{
			{IP: net.ParseIP("1.1.1.1")},
			{IP: net.ParseIP("127.0.0.1")},
		}},
	}

	if _, err := tlsScanner.resolveTarget(context.Background(), "example.com", 443); err == nil {
		t.Fatal("expected mixed DNS answer set to be rejected")
	}
}

func TestIsPublicIPRejectsReservedDocumentationNetworks(t *testing.T) {
	for _, value := range []string{"192.0.2.10", "198.51.100.2", "203.0.113.5", "2001:db8::1"} {
		if isPublicIP(net.ParseIP(value)) {
			t.Fatalf("expected %s to be rejected", value)
		}
	}
	if !isPublicIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("expected 1.1.1.1 to be accepted")
	}
}

func TestHSTSProbeUsesPinnedAddressAndOriginalHost(t *testing.T) {
	var observedHost string
	var observedSNI string
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		observedHost = request.Host
		observedSNI = request.TLS.ServerName
		writer.Header().Set("Strict-Transport-Security", "max-age=31536000")
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	target := resolvedTarget{
		domain:  "example.com",
		port:    443,
		ip:      net.ParseIP("127.0.0.1"),
		address: server.Listener.Addr().String(),
	}
	tlsScanner := &Scanner{timeout: 2 * time.Second, resolver: net.DefaultResolver}

	header, checked := tlsScanner.checkHSTS(context.Background(), target)
	if !checked {
		t.Fatal("expected HSTS probe to complete")
	}
	if header != "max-age=31536000" {
		t.Fatalf("unexpected HSTS header %q", header)
	}
	if observedHost != "example.com" {
		t.Fatalf("expected original HTTP host, got %q", observedHost)
	}
	if observedSNI != "example.com" {
		t.Fatalf("expected original SNI hostname, got %q", observedSNI)
	}
}
