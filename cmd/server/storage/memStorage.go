package storage

type MemStorage struct {
	MetricTypes map[string]*MetricType
}

type MetricType struct {
	Gauge   float64
	Counter int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		MetricTypes: make(map[string]*MetricType),
	}
}

func (m *MemStorage) UpdateMetric(metricName, typeMetric string, value interface{}) {

	if _, ok := m.MetricTypes[metricName]; !ok {
		m.MetricTypes[metricName] = &MetricType{
			Counter: 0,
			Gauge:   0,
		}
	}
	if typeMetric == "gauge" {
		m.MetricTypes[metricName].Gauge = value.(float64)
	} else {
		m.MetricTypes[metricName].Counter += value.(int64)
	}
}
