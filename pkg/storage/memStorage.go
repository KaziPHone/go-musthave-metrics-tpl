package storage

import (
	"fmt"

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
}

type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

func NewMemStorage() IStorage {
	return &MStorage{
		MetricTypes: make(map[string]*MetricType),
	}
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
	return nil
}

func (m *MStorage) ListMetrics() map[string]*MetricType {
	return m.MetricTypes
}

func (m *MStorage) GetMetric(metricName string) (*MetricType, bool) {
	metric, found := m.MetricTypes[metricName]
	return metric, found
}
