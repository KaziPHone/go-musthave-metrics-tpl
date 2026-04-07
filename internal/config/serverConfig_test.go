package config

import (
	"os"
	"testing"
)

// TestNewConfigServer_Defaults проверяет, что NewConfigServer возвращает значения по умолчанию
func TestNewConfigServer_Defaults(t *testing.T) {
	// сохраняем и восстанавливаем os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cmd"}

	cfg, err := NewConfigServer()
	if err != nil {
		t.Fatalf("NewConfigServer returned error: %v", err)
	}
	if cfg.Host == "" {
		t.Fatalf("expected default host to be set")
	}
}

// TestNewConfigServer_FromJSON проверяет чтение конфигурации из JSON-файла
func TestNewConfigServer_FromJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cfg-*.json")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{"address":"127.0.0.1:9999","trusted_subnet":"10.10.0.0/16"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	tmpFile.Close()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-c", tmpFile.Name()}

	cfg, err := NewConfigServer()
	if err != nil {
		t.Fatalf("NewConfigServer returned error: %v", err)
	}
	if cfg.Host != "127.0.0.1:9999" {
		t.Fatalf("expected host from json, got %s", cfg.Host)
	}
	if cfg.TrustedSubnet != "10.10.0.0/16" {
		t.Fatalf("expected trusted_subnet from json, got %s", cfg.TrustedSubnet)
	}
}
