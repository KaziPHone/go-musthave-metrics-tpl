package config

import (
	"os"
	"testing"
)

func TestNewConfigAgent_Defaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	cfg, err := NewConfigAgent()
	if err != nil {
		t.Fatalf("NewConfigAgent error: %v", err)
	}
	if cfg.Host == "" {
		t.Fatalf("expected default host set")
	}
}

func TestNewConfigAgent_FromJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "agentcfg-*.json")
	if err != nil {
		t.Fatalf("tmpfile: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `{"address":"localhost:7777","grpc_address":"127.0.0.1:50051"}`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	tmpFile.Close()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-c", tmpFile.Name()}

	cfg, err := NewConfigAgent()
	if err != nil {
		t.Fatalf("NewConfigAgent error: %v", err)
	}
	if cfg.Host != "localhost:7777" {
		t.Fatalf("expected host from json, got %s", cfg.Host)
	}
	if cfg.GRPCAddress != "127.0.0.1:50051" {
		t.Fatalf("expected grpc from json, got %s", cfg.GRPCAddress)
	}

}
