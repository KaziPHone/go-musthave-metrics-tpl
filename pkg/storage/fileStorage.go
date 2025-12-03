package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/rs/zerolog/log"
)

// initStorageFile инициализирует файловое хранилище
func (m *MStorage) initStorageFile() {
	if !m.isStorageFile() {
		return
	}
	log.Info().Msg("use file storage")
	m.loadStorageFile()
	if m.storeInterval > 0 {
		go m.storageFileTicker()
	}
}

// loadStorageFile загружает метрики из файла
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

// storageFileTicker запускает таймер для сохранения метрик в файл
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

// saveStorageMetrics сохраняет метрики в файл
func (m *MStorage) saveStorageMetrics() {

	if m.dataBase.dataBaseDsn != "" {
		return
	}

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
}
