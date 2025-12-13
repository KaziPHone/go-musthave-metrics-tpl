package agent

import (
	"sync"
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
		PollInterval:   3, // каждые 3 секунды
	}

	agent := NewAgent(cfg)
	stopCh := make(chan struct{})

	// 🔥 Добавим WaitGroup, чтобы дождаться завершения горутины
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done() // ✅ Указываем, что горутина завершена
		agent.monitoringMetrics(stopCh)
	}()

	// Ждём ~2 опроса (PollInterval=3, значит, за 4+ секунд — 1 или 2 раза)
	time.Sleep(4 * time.Second)

	// Останавливаем
	close(stopCh)

	// 🔥 Ждём, пока горутина полностью завершит работу
	wg.Wait()

	// Теперь безопасно читаем, потому что запись остановлена

	t.Run("Expected keys in metrics", func(t *testing.T) {
		// 🔒 Защитим чтение мьютексом — на всякий случай (если в будущем изменится поведение)
		agent.mu.Lock()
		for _, key := range expectedKeys {
			if _, ok := agent.metrics[key]; !ok {
				t.Errorf("Expected key %s not found in metrics", key)
			}
		}
		agent.mu.Unlock()
	})

	t.Run("Expected pollCount to be incremented", func(t *testing.T) {
		// 🔒 Защитим и pollCount (или сделаем его atomic, см. ниже)
		agent.mu.Lock()
		if agent.pollCount != 2 {
			t.Errorf("Expected pollCount to be 2, got %d", agent.pollCount)
		}
		agent.mu.Unlock()
	})
}
