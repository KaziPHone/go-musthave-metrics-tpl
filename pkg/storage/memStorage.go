package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/rs/zerolog/log"
)

type IStorage interface {
	UpdateMetric(metricName, typeMetric string, value interface{}) error
	ListMetrics() map[string]*MetricType
	GetMetric(metricName string) (*MetricType, bool)
	GracefulStop(server *http.Server)
}

type MStorage struct {
	MetricTypes    map[string]*MetricType
	fileStorage    string `env:"FILE_STORAGE_PATH"`
	storeInterval  int    `env:"STORE_INTERVAL"`
	restore        bool   `env:"RESTORE"`
}

type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

func NewMemStorage(cfg config.ServerConfig) IStorage {
	storage := &MStorage{
		MetricTypes:    make(map[string]*MetricType),
		fileStorage:    cfg.FileStorage,
		storeInterval:  cfg.StoreInterval,
		restore:        cfg.Restore,
	}
	storage.loadStorageFile()
	if storage.storeInterval > 0 {
		go storage.storageFileTicker()
	}

	return storage
}

func (m *MStorage) UpdateMetric(metricName, typeMetric string, value interface{}) error {

	if value == nil {
		return fmt.Errorf("value is nil")
	}

	if _, ok := m.MetricTypes[metricName]; !ok {
		m.MetricTypes[metricName] = &MetricType{
			Counter: 0,
			Gauge:   0,
		}
	}
	if typeMetric == models.Gauge {
		switch v := value.(type) {
		case float64:
			m.MetricTypes[metricName].Gauge = v
		case *float64:
			if v == nil {
				return fmt.Errorf("value is nil")
			}
			m.MetricTypes[metricName].Gauge = *v
		default:
			return fmt.Errorf("unexpected type for gauge: %T", value)
		}
		m.MetricTypes[metricName].Mtype = typeMetric
	} else {
		switch v := value.(type) {
		case float64:
			m.MetricTypes[metricName].Counter += int64(v)
		case *float64:
			if v == nil {
				return fmt.Errorf("value is nil")
			}
			m.MetricTypes[metricName].Counter += int64(*v)
		case int64:
			m.MetricTypes[metricName].Counter += v
		case *int64:
			if v == nil {
				return fmt.Errorf("value is nil")
			}
			m.MetricTypes[metricName].Counter += *v
		default:
			return fmt.Errorf("unexpected type for counter: %T", value)
		}
		m.MetricTypes[metricName].Mtype = typeMetric
	}

	if m.storeInterval == 0 {
		m.saveStorageMetrics()
	}

	return nil
}

func (m *MStorage) ListMetrics() map[string]*MetricType {
	return m.MetricTypes
}

func (m *MStorage) GetMetric(metricName string) (*MetricType, bool) {
	metric, found := m.MetricTypes[metricName]
	return metric, found
}

func (m *MStorage) loadStorageFile() error {

	metricsStorage := make([]models.Metrics, 0)

	if m.restore {

		data, err := os.ReadFile(m.fileStorage)
		if err != nil {
			return fmt.Errorf("не удалось прочитать файл %s: %v", m.fileStorage, err)
		}
		err = json.Unmarshal(data, &metricsStorage)
		if err != nil {
			return fmt.Errorf("ошибка десериализации JSON: %v", err)
		}

		for _, metric := range metricsStorage {
			m.MetricTypes[metric.ID] = &MetricType{
				Mtype: metric.MType,
			}
			if metric.MType == models.Gauge {
				if metric.Value != nil {
					m.MetricTypes[metric.ID].Gauge = *metric.Value
				}
			} else {
				if metric.Delta != nil {
					m.MetricTypes[metric.ID].Counter = int64(*metric.Delta)
				}
			}
		}

	}
	return nil
}

func (m *MStorage) storageFileTicker() {

	if m.storeInterval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(m.storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.saveStorageMetrics()
	}
}

func (m *MStorage) saveStorageMetrics() {

	metrics := make([]models.Metrics, 0)
	for id, metric := range m.MetricTypes {
		if metric.Mtype == models.Gauge {
			metrics = append(metrics, models.Metrics{
				ID:    id,
				MType: metric.Mtype,
				Value: &metric.Gauge,
			})
		} else {
			metrics = append(metrics, models.Metrics{
				ID:    id,
				MType: metric.Mtype,
				Delta: &metric.Counter,
			})
		}
	}

	dir, _ := filepath.Split(m.fileStorage)

	if _, err := os.Stat(dir); os.IsNotExist(err) && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			 log.Err(err)
		}
	}

	jsonData, err := json.MarshalIndent(metrics, "", "\t")
	if err != nil {
		log.Err(err)
	}

	err = os.WriteFile(m.fileStorage, jsonData, 0644)
	if err != nil {
		log.Err(err)
	}

	fmt.Println("save to file")
}

func (m *MStorage) GracefulStop(server *http.Server) {
	stopChan := make(chan os.Signal, 1)
    signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
        <-stopChan
		m.saveStorageMetrics()
        log.Print("Signal received, initiating save storage and graceful shutdown ...")
		os.Exit(0)
	}()
}