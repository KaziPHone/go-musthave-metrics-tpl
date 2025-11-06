package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_NormalCase(t *testing.T) {
	os.Setenv("ADDRESS", "localhost:8081")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("POLL_INTERVAL", "5")

	defer resetEnvironmentVars()

	cfg, err := NewConfigAgent()
	require.NoError(t, err)

	expectedCfg := &AgentConfig{
		Host:           "localhost:8081",
		ReportInterval: 15,
		PollInterval:   5,
	}

	assert.Equal(t, expectedCfg, cfg)
}

func TestNewConfig_DefaultValues(t *testing.T) {

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")

	defer resetEnvironmentVars()

	cfg, err := NewConfigAgent()
	require.NoError(t, err)

	expectedCfg := &AgentConfig{
		Host:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   3,
	}

	assert.Equal(t, expectedCfg, cfg)
}

// func TestNewConfig_InvalidFlag(t *testing.T) {

// 	os.Args = []string{
// 		"program_name", "--v=a",
// 	}

// 	_, err := NewConfig()

// 	// Сообщение об ошибке содержит предупреждение о некорректном значении
// 	assert.Contains(t, err.Error(), "переданы неизвестные флаги")
// }

// Вспомогательная функция для сброса переменных окружения
func resetEnvironmentVars() {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
}
