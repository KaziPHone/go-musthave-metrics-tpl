package storage

import "errors"

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
		if val, ok := value.(float64); !ok {
			return errors.New("expected float64 for gauge metric")
		} else {
			m.MetricTypes[metricName].Gauge = val
		}

	} else {
		if val, ok := value.(int64); !ok {
			return errors.New("expected int64 for counter metric")
		} else {
			m.MetricTypes[metricName].Counter += val
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
