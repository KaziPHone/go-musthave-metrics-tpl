package storage

import (
	"testing"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func TestUpdateMetricMemoryAndGetters(t *testing.T) {
	m := &MStorage{
		MetricTypes: make(map[string]*MetricType),
		dataBase:    newDataBase("", ""),
	}

	m.UpdateMetric("g1", models.Gauge, 2.5)
	mt, ok := m.GetMetric("g1")
	if !ok || mt.Gauge != 2.5 {
		t.Fatalf("expected g1 gauge 2.5, got %+v ok=%v", mt, ok)
	}

	m.UpdateMetric("c1", models.Counter, int64(3))
	mtc, ok := m.GetMetric("c1")
	if !ok || mtc.Counter != 3 {
		t.Fatalf("expected c1 counter 3, got %+v ok=%v", mtc, ok)
	}

	v := 7.7
	metrics := []models.Metrics{{ID: "g2", MType: models.Gauge, Value: &v}}
	m.UpdatesMetrics(metrics)
	if mg, ok := m.GetMetric("g2"); !ok || mg.Gauge != v {
		t.Fatalf("expected g2 gauge %v, got %+v", v, mg)
	}

	list := m.ListMetrics()
	if len(list) < 3 {
		t.Fatalf("expected at least 3 metrics, got %d", len(list))
	}
}
