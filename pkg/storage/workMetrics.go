// Package storage предоставляет реализации методов работы с метриками.
package storage

import models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"

// UpdateMetric обновляет одну метрику по имени.
//
// Параметры:
//   - metricName: имя метрики
//   - typeMetric: тип метрики ("gauge" или "counter")
//   - value: значение метрики (float64 для gauge, int64 для counter)
//
// Поведение зависит от типа хранилища:
//   - Database: сохраняет в PostgreSQL
//   - File: обновляет память + сохраняет в файл (если StoreInterval == 0)
//   - Memory: обновляет только в памяти
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

// ListMetrics возвращает все метрики из хранилища.
//
// Возвращает:
//   - map[string]*MetricType: карта метрик по имени
//
// При работе с базой данных вызывает getMetrics() из dataBase.
func (m *MStorage) ListMetrics() map[string]*MetricType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.isStorageBD() {
		return m.dataBase.getMetrics()
	}
	return m.MetricTypes
}

// GetMetric возвращает метрику по имени.
//
// Параметры:
//   - metricName: имя метрики
//
// Возвращает:
//   - *MetricType: метрика или nil если не найдена
//   - bool: true если метрика найдена
func (m *MStorage) GetMetric(metricName string) (*MetricType, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.isStorageBD() {
		return m.dataBase.getMetric(metricName)
	}
	metric, found := m.MetricTypes[metricName]
	return metric, found
}

// UpdatesMetrics обновляет несколько метрик за один вызов.
//
// Параметры:
//   - metrics: срез метрик для обновления
//
// Метод последовательно вызывает UpdateMetric для каждой метрики,
// определяя тип (gauge/counter) и значение (Value/Delta).
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
