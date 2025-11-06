package storage

import (
	"fmt"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)



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