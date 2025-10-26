package agent

import (
	"testing"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

func TestMonitoringMetrics(t *testing.T) {

	expectedKeys := []string{
		"Alloc", "BuckHashSys", "GCCPUFraction", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
		"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
		"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue",
	}

	cfg := config.AgentConfig{
		Host:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   3,
	}

	agent := NewAgent(cfg)
	stopCh := make(chan struct{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		agent.monitoringMetrics(stopCh)
	}()

	time.Sleep(4 * time.Second)
	close(stopCh)

	t.Run("Expected keys in metrics", func(t *testing.T) {
		for _, key := range expectedKeys {
			if _, ok := agent.metrics[key]; !ok {
				t.Errorf("Expected key %s not found in metrics", key)
			}
		}
	})

	t.Run("Expected pollCount to be incremented", func(t *testing.T) {
		if agent.pollCount != 1 {
			t.Errorf(
				"Expected pollCount to be 3, got %d",
				agent.pollCount,
			)
		}
	})

}
