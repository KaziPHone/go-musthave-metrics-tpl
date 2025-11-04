package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

type IStorage interface {
	UpdateMetric(metricName, typeMetric string, value interface{}) error
	ListMetrics() map[string]*MetricType
	GetMetric(metricName string) (*MetricType, bool)
	UpdateMetricV2(metric models.Metrics) error
}

type MStorage struct {
	MetricTypes map[string]*MetricType
	fileStorage string `env:"FILE_STORAGE_PATH"`
	storeInterval int `env:"STORE_INTERVAL"`
	restore bool `env:"RESTORE"`
	metricsStorage []models.Metrics
}

type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

func NewMemStorage(cfg config.ServerConfig) IStorage {
	storage := &MStorage{
		MetricTypes: make(map[string]*MetricType),
		fileStorage: cfg.FileStorage,
		storeInterval: cfg.StoreInterval,
		restore: cfg.Restore,
		metricsStorage: []models.Metrics{},
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
	return nil
}

// UpdateMetricV2 - обновление метрик по /update через json
func (m *MStorage) UpdateMetricV2(metric models.Metrics) error {
	if _, ok := m.MetricTypes[metric.ID]; !ok {
		m.MetricTypes[metric.ID] = &MetricType{}
	}

	if metric.MType == models.Gauge {
		if metric.Value == nil {
			return fmt.Errorf("value is nil")
		}
		m.MetricTypes[metric.ID].Gauge = *metric.Value

	} else {
		if metric.Delta == nil {
			return fmt.Errorf("value is nil")
		}
		m.MetricTypes[metric.ID].Counter += int64(*metric.Delta)
	}
	m.MetricTypes[metric.ID].Mtype = metric.MType

	m.updateMetricStorage(metric)
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

// loadStorageFile - загрузка метрик из файла
func (m *MStorage) loadStorageFile() error {
	
	if m.restore {

		data, err := ioutil.ReadFile(m.fileStorage)
		if err != nil {
			return fmt.Errorf("не удалось прочитать файл %s: %v", m.fileStorage, err)
		}
		err = json.Unmarshal(data, &m.metricsStorage)
		if err != nil {
			return fmt.Errorf("ошибка десериализации JSON: %v", err)
		}

	}
	return nil
}

func (m *MStorage) storageFileTicker() {

	if m.storeInterval <= 0 {
        fmt.Println("Синхронное сохранение включено.")
        return
    }
	fmt.Println("start ticker", m.storeInterval)
    ticker := time.NewTicker(time.Duration(m.storeInterval) * time.Second)
    defer ticker.Stop()
	

	for range ticker.C {
		fmt.Println("saver ")
        m.saveStorageMetrics()
    }
}

func (m *MStorage) updateMetricStorage(metric models.Metrics) {

	if len(m.metricsStorage) == 0 {
		m.metricsStorage = append(m.metricsStorage, metric)
		return
	}

	found := false
	for i := range m.metricsStorage {
		if m.metricsStorage[i].ID == metric.ID &&  m.metricsStorage[i].MType == metric.MType {
			found = true
			if  m.metricsStorage[i].MType == models.Gauge {
				m.metricsStorage[i].Value = metric.Value
			} else {
				*m.metricsStorage[i].Delta += *metric.Delta
			}
		}
		
	}
	if !found {
		m.metricsStorage = append(m.metricsStorage, metric)
	}
}

func (m *MStorage) saveStorageMetrics() error {

	jsonData, err := json.MarshalIndent(m.metricsStorage, "", "\t")
	fmt.Println(err)

	err = ioutil.WriteFile(m.fileStorage, jsonData, 0644)
	fmt.Println(err)

	fmt.Printf("Сохранились метрики в %s.\n", m.fileStorage)

	return nil
}