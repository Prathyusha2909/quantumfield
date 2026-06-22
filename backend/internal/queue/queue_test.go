package queue

import "testing"

func TestScanJobEncodingRoundTrip(t *testing.T) {
	expected := ScanJob{
		ScanID:  "scan-123",
		AssetID: "asset-456",
		UserID:  "user-789",
	}
	payload, err := encodeScanJob(expected)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := decodeScanJob(payload)
	if err != nil {
		t.Fatal(err)
	}
	if *actual != expected {
		t.Fatalf("unexpected job after round trip: %+v", actual)
	}
}

func TestScanJobDecoderRejectsIncompletePayload(t *testing.T) {
	if _, err := decodeScanJob([]byte(`{"scan_id":"scan-123"}`)); err == nil {
		t.Fatal("expected incomplete queue payload to be rejected")
	}
	if _, err := decodeScanJob([]byte(`not-json`)); err == nil {
		t.Fatal("expected malformed queue payload to be rejected")
	}
}
