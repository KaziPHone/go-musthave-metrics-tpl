package agent

import (
	"bytes"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	pollInterval   int
	reportInterval int
	url            string
	pollCount      int32
	metrics        map[string]float64
	mu             sync.Mutex
}

func NewAgent(pollInterval, reportInterval int, Host string) *Agent {
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		url:            "http://" + Host + "/update",
		pollCount:      0,
		metrics:        make(map[string]float64),
		mu:             sync.Mutex{},
	}
}

func (a *Agent) sendRequest(typeMetric, metricName string, value string) {

	resp, err := http.Post(a.url+"/"+typeMetric+"/"+metricName+"/"+value, "text/plain", bytes.NewBuffer(nil))
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (a *Agent) reportMetrics() {

	for {
		a.mu.Lock()
		for key, value := range a.metrics {
			a.sendRequest("gauge", key, strconv.FormatFloat(value, 'f', -1, 64))
		}
		a.sendRequest("counter", "PollCount", strconv.Itoa(int(a.pollCount)))
		a.mu.Unlock()
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
