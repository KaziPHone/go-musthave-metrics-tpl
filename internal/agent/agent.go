// agent/agent.go
package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"

	//"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Agent struct {
	pollInterval   int
	reportInterval int
	url            string
	pollCount      int32
	metrics        map[string]float64
	mu             sync.Mutex
	httpClient     *http.Client
	maxRetries     int
	retryDelays    []time.Duration
	key            string
	rateLimit      int
}

func NewAgent(cfg config.AgentConfig) *Agent {
	return &Agent{
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		key:            cfg.Key,
		url:            "http://" + cfg.Host + "/updates/",
		pollCount:      0,
		metrics:        make(map[string]float64),
		mu:             sync.Mutex{},
		httpClient:     &http.Client{},
		maxRetries:     3,
		retryDelays:    []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
		rateLimit:      cfg.RateLimit, // Должно быть из флага / env
	}
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("error writing to gzip: %v", err)
	}
	if err := gw.Flush(); err != nil {
		return nil, fmt.Errorf("the flush gzip error: %v", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip closing error: %v", err)
	}
	return buf.Bytes(), nil
}

// sendMetric отправляет один пакет метрик (реализует retry-логику)
func (a *Agent) sendMetric(metrics []models.Metrics) {
	for attempt := 0; ; attempt++ {
		out, err := json.Marshal(metrics)
		if err != nil {
			fmt.Printf("Error marshalling: %v\n", err)
			return
		}

		compressedData, err := compress(out)
		if err != nil {
			fmt.Printf("Error compress: %v\n", err)
			return
		}

		req, err := http.NewRequest("POST", a.url, bytes.NewReader(compressedData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		if a.key != "" {
			req.Header.Set("HashSHA256", helpers.CalcSHA256Hash(compressedData))
		}

		resp, err := a.httpClient.Do(req)
		if err != nil {
			if attempt >= a.maxRetries {
				fmt.Printf("Unsuccessful request after max retries: %v\n", err)
				return
			}
			time.Sleep(a.retryDelays[attempt])
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if attempt >= a.maxRetries {
				fmt.Printf("Failed with status %d\n", resp.StatusCode)
				return
			}
			time.Sleep(a.retryDelays[attempt])
			continue
		}

		fmt.Printf("Successfully sent %d metrics\n", len(metrics))
		return
	}
}

func (a *Agent) worker(jobs <-chan []models.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for metrics := range jobs {
		a.sendMetric(metrics)
	}
}

func (a *Agent) collectSystemMetrics() {
	v, _ := mem.VirtualMemory()
	a.mu.Lock()
	a.metrics["TotalMemory"] = float64(v.Total)
	a.metrics["FreeMemory"] = float64(v.Free)
	a.mu.Unlock()

	// CPU Utilization
	percents, err := cpu.Percent(time.Second, true)
	if err == nil {
		a.mu.Lock()
		for i, p := range percents {
			a.metrics[fmt.Sprintf("CPUUtilization%d", i+1)] = p
		}
		a.mu.Unlock()
	}
}

func (a *Agent) collectRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.mu.Lock()
	a.metrics["Alloc"] = float64(memStats.Alloc)
	a.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	a.metrics["GCCPUFraction"] = memStats.GCCPUFraction
	a.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	a.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	a.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	a.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	a.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	a.metrics["HeapSys"] = float64(memStats.HeapSys)
	a.metrics["LastGC"] = float64(memStats.LastGC)
	a.metrics["Lookups"] = float64(memStats.Lookups)
	a.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	a.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	a.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	a.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	a.metrics["Mallocs"] = float64(memStats.Mallocs)
	a.metrics["NextGC"] = float64(memStats.NextGC)
	a.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	a.metrics["NumGC"] = float64(memStats.NumGC)
	a.metrics["OtherSys"] = float64(memStats.OtherSys)
	a.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	a.metrics["StackInuse"] = float64(memStats.StackInuse)
	a.metrics["StackSys"] = float64(memStats.StackSys)
	a.metrics["Sys"] = float64(memStats.Sys)
	a.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	a.metrics["RandomValue"] = rand.Float64()
	a.metrics["Frees"] = float64(memStats.Frees)
	a.metrics["GCSys"] = float64(memStats.GCSys)
	a.pollCount++
	a.mu.Unlock()
}

func (a *Agent) pollMetrics(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			a.collectRuntimeMetrics()
			a.collectSystemMetrics()
		}
	}
}

func (a *Agent) reportMetrics(stopCh <-chan struct{}) {
	jobs := make(chan []models.Metrics, a.rateLimit*2)
	var wg sync.WaitGroup

	for i := 0; i < a.rateLimit; i++ {
		wg.Add(1)
		go a.worker(jobs, &wg)
	}

	ticker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			close(jobs)
			wg.Wait()
			return
		case <-ticker.C:
			metrics := a.copyMetrics()
			if len(metrics) > 0 {
				select {
				case jobs <- metrics:
				default:
					fmt.Println("Rate limit exceeded, skipping batch")
				}
			}
		}
	}
}

func (a *Agent) copyMetrics() []models.Metrics {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]models.Metrics, 0, len(a.metrics)+1)

	for key, value := range a.metrics {
		result = append(result, models.Metrics{
			ID:    key,
			MType: models.Gauge,
			Value: &value,
		})
	}

	v := int64(a.pollCount)
	result = append(result, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &v,
	})

	return result
}

func (a *Agent) Start(stopCh <-chan struct{}) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.pollMetrics(stopCh)
	}()

	go func() {
		defer wg.Done()
		a.reportMetrics(stopCh)
	}()

	wg.Wait()
}
