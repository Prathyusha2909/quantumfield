package target

import "testing"

func TestNormalize(t *testing.T) {
	domain, port, err := Normalize("https://Example.COM:8443/path", 0)
	if err != nil {
		t.Fatal(err)
	}
	if domain != "example.com" || port != 8443 {
		t.Fatalf("unexpected target %s:%d", domain, port)
	}
}

func TestNormalizeRejectsSingleLabel(t *testing.T) {
	if _, _, err := Normalize("localhost", 443); err == nil {
		t.Fatal("expected validation error")
	}
}
