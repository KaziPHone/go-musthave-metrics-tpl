package storage

import models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"

// UpdateMetric обновление 1й метрики от агента
func (m *MStorage) UpdateMetric(metricName, typeMetric string, value interface{}) {

	switch {
	case m.isStorageBD():
		m.dataBase.insertMetric(metricName, typeMetric, value)
	case m.isStorageFile():
		m.updateMetricMemory(metricName, typeMetric, value)
		if m.storeInterval == 0 {
			m.saveStorageMetrics()
		}
	default:
		m.updateMetricMemory(metricName, typeMetric, value)
	}

}

// ListMetrics возвращает список метрик
func (m *MStorage) ListMetrics() map[string]*MetricType {
	if m.isStorageBD() {
		return m.dataBase.getMetrics()
	}
	return m.MetricTypes
}

// GetMetric возвращает метрику по имени
func (m *MStorage) GetMetric(metricName string) (*MetricType, bool) {
	if m.isStorageBD() {
		return m.dataBase.getMetric(metricName)
	}
	metric, found := m.MetricTypes[metricName]
	return metric, found
}

// UpdatesMetrics обновление всех метрик
func (m *MStorage) UpdatesMetrics(metrics []models.Metrics) {
	// Предвычисляем количество итераций
	n := len(metrics)
	for i := 0; i < n; i++ {
		metric := metrics[i]
		var v interface{}
		if metric.MType == models.Gauge {
			v = metric.Value
		} else {
			v = metric.Delta
		}
		m.UpdateMetric(metric.ID, metric.MType, v)
	}
}