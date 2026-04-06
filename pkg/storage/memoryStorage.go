package storage

import (
	"fmt"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

// updateMetricMemory обновление 1й метрики (внутренняя функция, с блокировкой)
func (m *MStorage) updateMetricMemory(metricName, typeMetric string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Используем value directly without nested pointer checks where possible
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
			if v != nil {
				m.MetricTypes[metricName].Gauge = *v
			}
		default:
			return fmt.Errorf("unexpected type for gauge: %T", value)
		}
		m.MetricTypes[metricName].Mtype = typeMetric
	} else {
		switch v := value.(type) {
		case float64:
			m.MetricTypes[metricName].Counter += int64(v)
		case *float64:
			if v != nil {
				m.MetricTypes[metricName].Counter += int64(*v)
			}
		case int64:
			m.MetricTypes[metricName].Counter += v
		case *int64:
			if v != nil {
				m.MetricTypes[metricName].Counter += *v
			}
		default:
			return fmt.Errorf("unexpected type for counter: %T", value)
		}
		m.MetricTypes[metricName].Mtype = typeMetric
	}

	return nil
}
