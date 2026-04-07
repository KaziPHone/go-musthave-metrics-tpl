package helpers

import (
	"bytes"
	"testing"
)

func TestCalcSHA256HashAndIsBad(t *testing.T) {
	data := []byte("hello world")
	h := CalcSHA256Hash(data)
	if h == "" {
		t.Fatalf("expected non-empty hash")
	}

	if IsBadShaRequest(data, h) {
		t.Fatalf("expected IsBadShaRequest to be false for matching hash")
	}

	if IsBadShaRequest(data, "") {
		t.Fatalf("expected IsBadShaRequest to be false for empty hash")
	}

	if !IsBadShaRequest(data, "deadbeef") {
		t.Fatalf("expected IsBadShaRequest to be true for non-matching hash")
	}
}

func TestCalcSHA256HashBuffer(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("buffered data")
	h := CalcSHA256HashBuffer(buf)
	if h == "" {
		t.Fatalf("expected non-empty hash from buffer")
	}
}
