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
)

type Agent struct {
	pollInterval   int
	reportInterval int
	url            string
	pollCount      int32
	metrics        map[string]float64
	mu             sync.Mutex
	httpClient     *http.Client
}

func NewAgent(cfg config.AgentConfig) *Agent {
	return &Agent{
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		url:            "http://" + cfg.Host + "/updates/",
		pollCount:      0,
		metrics:        make(map[string]float64),
		mu:             sync.Mutex{},
		httpClient:     &http.Client{},
	}
}

func compress(data []byte) ([]byte, error) {

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка записи в gzip: %v", err)
	}
	err = gw.Flush()
	if err != nil {
		return nil, fmt.Errorf("ошибка flush gzip: %v", err)
	}
	err = gw.Close()
	if err != nil {
		return nil, fmt.Errorf("ошибка закрытия gzip: %v", err)
	}
	return buf.Bytes(), nil
}

func (a *Agent) sendRequest(metrics []models.Metrics) {

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

	resp, err := a.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	resp.Body.Close()
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
			a.pollCount += 1
			a.mu.Unlock()
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
