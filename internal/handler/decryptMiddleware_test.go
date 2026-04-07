package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cryptoPkg "github.com/KaziPHone/go-musthave-metrics-tpl/pkg/crypto"
)

func TestDecryptRequestMiddleware_NilKey_PassesThrough(t *testing.T) {
	mw := DecryptRequestMiddleware(nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("plain")))
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if rr.Body.String() != "plain" {
		t.Fatalf("expected body to pass through, got %q", rr.Body.String())
	}
}

func TestDecryptRequestMiddleware_Decrypts(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pub := &priv.PublicKey

	plain := []byte("secret-data")
	enc, err := cryptoPkg.Encrypt(pub, plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	mw := DecryptRequestMiddleware(priv)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Write(b)
	})

	req := httptest.NewRequest("POST", "/", bytes.NewReader(enc))
	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if rr.Body.String() != string(plain) {
		t.Fatalf("expected decrypted body %q, got %q", string(plain), rr.Body.String())
	}
}
