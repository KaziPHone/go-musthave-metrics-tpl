package storage

import (
    "encoding/json"
    "os"
    "testing"

    models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func TestSaveStorageMetricsCreatesFile(t *testing.T) {
    tmpFile := os.TempDir() + "/test_metrics.json"
    defer os.Remove(tmpFile)

    m := &MStorage{
        MetricTypes: map[string]*MetricType{
            "g1": {Mtype: models.Gauge, Gauge: 1.23},
            "c1": {Mtype: models.Counter, Counter: 5},
        },
        fileStorage: tmpFile,
        dataBase:    newDataBase("", ""),
    }

    m.saveStorageMetrics()

    // прочитать файл и проверить содержимое
    data, err := os.ReadFile(tmpFile)
    if err != nil {
        t.Fatalf("expected file to be written, err: %v", err)
    }

    var arr []models.Metrics
    if err := json.Unmarshal(data, &arr); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if len(arr) == 0 {
        t.Fatalf("expected metrics in file")
    }
}
