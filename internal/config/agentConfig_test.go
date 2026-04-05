package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigAgent_NormalCase(t *testing.T) {
	os.Setenv("AGENT_HOST", "localhost:8081")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("RATE_LIMIT", "1")

	defer resetEnvironmentVars()

	cfg, err := NewConfigAgent()
	require.NoError(t, err)

	expectedCfg := &AgentConfig{
		Host:           "localhost:8081",
		ReportInterval: 15,
		PollInterval:   5,
		RateLimit:      1,
	}

	assert.Equal(t, expectedCfg, cfg)
}

// Вспомогательная функция для сброса переменных окружения
func resetEnvironmentVars() {
	os.Unsetenv("AGENT_HOST")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("RATE_LIMIT")
}
