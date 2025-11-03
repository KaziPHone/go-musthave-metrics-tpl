package storage

import "fmt"

type IStorage interface {
	UpdateMetric(metricName, typeMetric string, value interface{}) error
	ListMetrics() map[string]*MetricType
	GetMetric(metricName string) (*MetricType, bool)
}

type MStorage struct {
	MetricTypes map[string]*MetricType
}

type MetricType struct {
	Gauge   float64
	Counter int64
}

func NewMemStorage() IStorage {
	return &MStorage{
		MetricTypes: make(map[string]*MetricType),
	}
}

func (m *MStorage) UpdateMetric(metricName, typeMetric string, value interface{}) error {

	if _, ok := m.MetricTypes[metricName]; !ok {
		m.MetricTypes[metricName] = &MetricType{
			Counter: 0,
			Gauge:   0,
		}
	}
	if typeMetric == "gauge" {
		switch v := value.(type) {
		case float64:
			m.MetricTypes[metricName].Gauge = v
		case *float64:
			m.MetricTypes[metricName].Gauge = *v
		default:
			return fmt.Errorf("unexpected type for gauge: %T", value)
		}

	} else {
		switch v := value.(type) {
		case int64:
			m.MetricTypes[metricName].Counter += v
		case *int64:
			m.MetricTypes[metricName].Counter += *v
		default:
			return fmt.Errorf("unexpected type for counter: %T", value)
		}
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
