package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
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
	}
}

func compress(data []byte) ([]byte, error) {

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error writing to gzip: %v", err)
	}
	err = gw.Flush()
	if err != nil {
		return nil, fmt.Errorf("the flush gzip error: %v", err)
	}
	err = gw.Close()
	if err != nil {
		return nil, fmt.Errorf("gzip closing error: %v", err)
	}
	return buf.Bytes(), nil
}

func (a *Agent) sendRequest(metrics []models.Metrics) {

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
				fmt.Printf("Unsuccessful request sending after maximum attempts (%d)\n", a.maxRetries)
				return
			}

			fmt.Printf("Request sending error: %v, repeat via %v\n", err, a.retryDelays[attempt])
			time.Sleep(a.retryDelays[attempt])

			continue
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			if attempt >= a.maxRetries {
				fmt.Printf("Failed request (code: %d, body: %q) after maximum attempts (%d)\n", resp.StatusCode, bodyBytes, a.maxRetries)
				return
			}
			fmt.Printf("Incorrect status received (%d): %q, repeat after %v\n", resp.StatusCode, bodyBytes, a.retryDelays[attempt])
			time.Sleep(a.retryDelays[attempt])
			continue
		}

		fmt.Printf("The request was sent successfully!\n")

		return
	}
}

func (a *Agent) reportMetrics() {

	for {

		metrics := make([]models.Metrics, 0)

		for key, value := range a.metrics {
			metrics = append(metrics, models.Metrics{
				ID:    key,
				MType: models.Gauge,
				Value: &value,
			})
		}

		v := int64(a.pollCount)
		metrics = append(metrics, models.Metrics{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &v,
		})

		a.sendRequest(metrics)

		time.Sleep(time.Duration(a.reportInterval) * time.Second)

	}

}

func (a *Agent) monitoringMetrics(stopCh <-chan struct{}) {
	var memStats runtime.MemStats

	for {
		select {
		case <-stopCh:
			return
		default:
			a.mu.Lock()
			defer a.mu.Unlock()
			runtime.ReadMemStats(&memStats)
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
			a.pollCount += 1
			time.Sleep(time.Duration(a.pollInterval) * time.Second)
		}

	}
}

func (a *Agent) Start(stopCh <-chan struct{}) {
	monitoringStop := make(chan struct{})
	go a.monitoringMetrics(monitoringStop)
	go a.reportMetrics()
	<-stopCh
	close(monitoringStop)
}
