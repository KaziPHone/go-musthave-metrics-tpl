package storage


func (m *MStorage) UpdateMetric(metricName, typeMetric string, value interface{}) error {
	switch {
	case m.useStorage == useBd:
		m.dataBase.insertMetric(metricName, typeMetric, value)
	case m.useStorage == useFileStorage:
		m.updateMetricMemory(metricName, typeMetric, value)
		if m.storeInterval == 0 {
			m.saveStorageMetrics()
		}
	default:
		m.updateMetricMemory(metricName, typeMetric, value)
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

func (m *MStorage) setStorage() {
	switch {
	case m.dataBase.dataBaseDsn != "":
		m.useStorage = useBd
	case m.fileStorage != "":
		m.useStorage = useFileStorage
	default:
		m.useStorage = useMemory
	}
}

func (m *MStorage) IsConnectedDB() bool {
	return m.dataBase.isConnected
}