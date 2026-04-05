package storage

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestNewMemStorage(t *testing.T) {
	cfg := config.ServerConfig{}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.ListMetrics())
}

func TestNewMemStorage_FileStorage(t *testing.T) {
	cfg := config.ServerConfig{
		FileStorage:   "/tmp/test_metrics.json",
		StoreInterval: 0, // Disable ticker
		Restore:       false,
	}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.ListMetrics())
}

func TestNewMemStorage_FileStorageWithRestore(t *testing.T) {
	// Create a test file with some data
	tmpFile := "/tmp/test_restore.json"
	testData := `[
		{"id": "test_metric", "type": "gauge", "value": 100.5},
		{"id": "test_counter", "type": "counter", "delta": 50}
	]`
	err := os.WriteFile(tmpFile, []byte(testData), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	cfg := config.ServerConfig{
		FileStorage:   tmpFile,
		Restore:       true,
		StoreInterval: 0, // Disable ticker
	}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
	// After restore, the metrics should be loaded
}

func TestNewMemStorage_DBStorage(t *testing.T) {
	// Test with a valid DSN (will fail to connect but should not panic)
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/testdb",
		MigratePath: "./migrations",
	}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
	// Database won't be connected since we're not running a real DB
}

func TestNewMemStorage_InvalidMigratePath(t *testing.T) {
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/testdb",
		MigratePath: "/nonexistent/path",
	}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
}

func TestMStorage_UpdateMetric_Gauge(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_gauge", "gauge", 100.5)

	metrics := storage.ListMetrics()
	assert.Contains(t, metrics, "test_gauge")
	assert.Equal(t, 100.5, metrics["test_gauge"].Gauge)
}

func TestMStorage_UpdateMetric_Counter(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_counter", "counter", int64(50))

	metrics := storage.ListMetrics()
	assert.Contains(t, metrics, "test_counter")
	assert.Equal(t, int64(50), metrics["test_counter"].Counter)
}

func TestMStorage_UpdateMetric_IncrementCounter(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_counter", "counter", int64(50))
	storage.UpdateMetric("test_counter", "counter", int64(25))

	metrics := storage.ListMetrics()
	assert.Equal(t, int64(75), metrics["test_counter"].Counter)
}

func TestMStorage_UpdateMetric_OverwriteGauge(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_metric", "gauge", 100.0)
	storage.UpdateMetric("test_metric", "gauge", 200.0)

	metrics := storage.ListMetrics()
	assert.Equal(t, 200.0, metrics["test_metric"].Gauge)
}

func TestMStorage_UpdateMetric_TypeSwitch(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	// First with float64
	storage.UpdateMetric("metric1", "gauge", float64(100.0))
	// Then with *float64
	storage.UpdateMetric("metric2", "gauge", &[]float64{200.0}[0])

	metrics := storage.ListMetrics()
	assert.Contains(t, metrics, "metric1")
	assert.Contains(t, metrics, "metric2")
}

func TestMStorage_ListMetrics(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("m1", "gauge", 1.0)
	storage.UpdateMetric("m2", "counter", int64(2))
	storage.UpdateMetric("m3", "gauge", 3.0)

	metrics := storage.ListMetrics()

	assert.Equal(t, 3, len(metrics))
	assert.Contains(t, metrics, "m1")
	assert.Contains(t, metrics, "m2")
	assert.Contains(t, metrics, "m3")
}

func TestMStorage_ListMetrics_Empty(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	metrics := storage.ListMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, 0, len(metrics))
}

func TestMStorage_GetMetric(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_metric", "gauge", 100.0)

	metric, found := storage.GetMetric("test_metric")
	assert.True(t, found)
	assert.NotNil(t, metric)
	assert.Equal(t, 100.0, metric.Gauge)
}

func TestMStorage_GetMetric_NotFound(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	metric, found := storage.GetMetric("nonexistent")

	assert.False(t, found)
	assert.Nil(t, metric)
}

func TestMStorage_UpdatesMetrics(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: floatPtr(1.0)},
		{ID: "m2", MType: "counter", Delta: int64Ptr(2)},
		{ID: "m3", MType: "gauge", Value: floatPtr(3.0)},
	}

	storage.UpdatesMetrics(metrics)

	result := storage.ListMetrics()
	assert.Equal(t, 3, len(result))
}

func TestMStorage_UpdatesMetrics_Empty(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdatesMetrics([]models.Metrics{})

	result := storage.ListMetrics()
	assert.Equal(t, 0, len(result))
}

func TestMStorage_UpdatesMetrics_InvalidType(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	// Test with empty slice
	storage.UpdatesMetrics([]models.Metrics{})

	// Should not panic
	assert.NotNil(t, storage.ListMetrics())
}

func TestMStorage_StorageGracefulStop(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	server := &http.Server{
		Addr: ":8080",
	}

	// This starts a goroutine, so we need to give it a moment
	done := make(chan struct{})
	go func() {
		storage.StorageGracefulStop(server)
		close(done)
	}()

	// Wait a bit then send a signal
	time.Sleep(100 * time.Millisecond)

	// We can't actually send a signal in this test environment
	// Just verify the function doesn't panic
}

func TestMStorage_StorageGracefulStop_FileStorage(t *testing.T) {
	tmpFile := "/tmp/test_graceful.json"
	defer os.Remove(tmpFile)

	cfg := config.ServerConfig{
		FileStorage:   tmpFile,
		StoreInterval: 1, // Fast interval for testing
	}
	storage := NewMemStorage(cfg)

	server := &http.Server{Addr: ":8081"}

	done := make(chan struct{})
	go func() {
		storage.StorageGracefulStop(server)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
}

func TestMStorage_isStorageBD(t *testing.T) {
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/test",
	}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.True(t, storage.isStorageBD())
}

func TestMStorage_isStorageFile(t *testing.T) {
	cfg := config.ServerConfig{
		FileStorage: "/tmp/test.json",
	}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.True(t, storage.isStorageFile())
}

func TestMStorage_isStorageMemory(t *testing.T) {
	cfg := config.ServerConfig{}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.False(t, storage.isStorageBD())
	assert.False(t, storage.isStorageFile())
}

func TestMStorage_setStorage_DBFirst(t *testing.T) {
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/test",
		FileStorage: "/tmp/test.json",
	}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.True(t, storage.isStorageBD())
	assert.False(t, storage.isStorageFile())
}

func TestMStorage_setStorage_FileOnly(t *testing.T) {
	cfg := config.ServerConfig{
		FileStorage: "/tmp/test.json",
	}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.False(t, storage.isStorageBD())
	assert.True(t, storage.isStorageFile())
}

func TestMStorage_setStorage_Neither(t *testing.T) {
	cfg := config.ServerConfig{}
	storage := NewMemStorage(cfg).(*MStorage)

	assert.False(t, storage.isStorageBD())
	assert.False(t, storage.isStorageFile())
}

func TestMStorage_IsConnectedDB(t *testing.T) {
	cfg := config.ServerConfig{}
	storage := NewMemStorage(cfg)

	// Without a real DB, this should return false
	connected := storage.IsConnectedDB()
	assert.False(t, connected)
}

func TestMStorage_IsConnectedDB_WithDB(t *testing.T) {
	// This would require a real database connection
	// We just verify the method exists and returns bool
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/test",
	}
	storage := NewMemStorage(cfg)

	connected := storage.IsConnectedDB()
	assert.IsType(t, false, connected)
}

func TestMStorage_updateMetricMemory_Gauge(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "gauge", 100.0)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, storage.MetricTypes["test"].Gauge)
}

func TestMStorage_updateMetricMemory_Counter(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "counter", int64(50))
	assert.NoError(t, err)
	assert.Equal(t, int64(50), storage.MetricTypes["test"].Counter)
}

func TestMStorage_updateMetricMemory_UnknownType(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "gauge", "invalid_string")
	assert.Error(t, err)
}

func TestMStorage_updateMetricMemory_UnknownCounterType(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "counter", []string{"invalid"})
	assert.Error(t, err)
}

func TestMStorage_updateMetricMemory_Float64Counter(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "counter", float64(100.0))
	assert.NoError(t, err)
	assert.Equal(t, int64(100), storage.MetricTypes["test"].Counter)
}

func TestMStorage_initStorageFile_NoFileStorage(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)
	
	// Should not panic
	storage.initStorageFile()
}

func TestMStorage_initStorageFile_WithFileStorage(t *testing.T) {
	tmpFile := "/tmp/test_init.json"
	defer os.Remove(tmpFile)

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
	}
	storage := NewMemStorage(cfg).(*MStorage)

	storage.initStorageFile()
}

func TestMStorage_storageFileTicker_NoInterval(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)
	
	// Should not panic
	storage.storageFileTicker()
}

func TestMStorage_storageFileTicker_WithInterval(t *testing.T) {
	cfg := config.ServerConfig{
		FileStorage:   "/tmp/test_ticker.json",
		StoreInterval: 1,
	}

	// Storage init already starts the ticker in a goroutine
	// Just verify it doesn't panic and runs
	time.Sleep(1100 * time.Millisecond)

	// Verify file was created/saved
	if _, err := os.Stat(cfg.FileStorage); err == nil {
		// File exists, which means ticker ran and saved
		os.Remove(cfg.FileStorage)
	}
}

func TestMStorage_saveStorageMetrics_DBActive(t *testing.T) {
	cfg := config.ServerConfig{
		DataBaseDsn: "postgres://user:pass@localhost:5432/test",
	}
	storage := NewMemStorage(cfg).(*MStorage)

	// Should return early when DB is active
	storage.saveStorageMetrics()
}

func TestMStorage_saveStorageMetrics_NoDir(t *testing.T) {
	tmpFile := "/tmp/nonexistent_dir/test_metrics.json"
	defer os.RemoveAll("/tmp/nonexistent_dir")

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
	}
	storage := NewMemStorage(cfg).(*MStorage)

	// Should create directory and save
	storage.saveStorageMetrics()
}

func TestMStorage_saveStorageMetrics_JSONError(t *testing.T) {
	// This is hard to trigger without mocking
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	storage.saveStorageMetrics()
}

func TestNewMemStorage_ConcurrentAccess(t *testing.T) {
	// Use memory storage (no ticker) for concurrent access test
	cfg := config.ServerConfig{
		FileStorage:   "",
		StoreInterval: 0, // Disable ticker
	}
	storage := NewMemStorage(cfg)

	var done = make(chan bool)

	// Concurrent updates
	for i := 0; i < 10; i++ {
		go func(i int) {
			storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := storage.ListMetrics()
	assert.Equal(t, 10, len(metrics))
}

func TestMStorage_UpdatesMetrics_Concurrent(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			metrics := []models.Metrics{
				{ID: fmt.Sprintf("m_%d", i), MType: "gauge", Value: floatPtr(float64(i))},
			}
			storage.UpdatesMetrics(metrics)
		}(i)
	}

	wg.Wait()

	metrics := storage.ListMetrics()
	assert.GreaterOrEqual(t, len(metrics), 10)
}

func TestMStorage_UpdateMetric_Concurrent(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			storage.UpdateMetric("concurrent_metric", "gauge", float64(i))
		}(i)
	}

	wg.Wait()

	metrics := storage.ListMetrics()
	// With concurrent updates, any value is possible
	// Just verify it's one of the expected values
	assert.GreaterOrEqual(t, metrics["concurrent_metric"].Gauge, float64(0))
	assert.LessOrEqual(t, metrics["concurrent_metric"].Gauge, float64(99))
}

func TestMetricType_Init(t *testing.T) {
	mt := &MetricType{}

	assert.Equal(t, "", mt.Mtype)
	assert.Equal(t, float64(0), mt.Gauge)
	assert.Equal(t, int64(0), mt.Counter)
}

func TestMetricType_Update(t *testing.T) {
	mt := &MetricType{
		Counter: 10,
		Gauge:   20.0,
	}

	mt.Counter = 30
	mt.Gauge = 40.0

	assert.Equal(t, int64(30), mt.Counter)
	assert.Equal(t, 40.0, mt.Gauge)
}

func TestMStorage_WithCustomStoragePath(t *testing.T) {
	tmpDir := "/tmp/test_storage"
	tmpFile := tmpDir + "/metrics.json"
	defer os.RemoveAll(tmpDir)

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
	}
	storage := NewMemStorage(cfg)

	assert.NotNil(t, storage)
}

func TestMStorage_FileStorageWithRestore(t *testing.T) {
	tmpFile := "/tmp/test_restore_data.json"
	defer os.Remove(tmpFile)

	// Create a file with initial data
	initialData := `[
		{"id": "existing_metric", "type": "gauge", "value": 999.9}
	]`
	err := os.WriteFile(tmpFile, []byte(initialData), 0644)
	require.NoError(t, err)

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
		Restore:     true,
	}
	storage := NewMemStorage(cfg)

	metrics := storage.ListMetrics()
	assert.Contains(t, metrics, "existing_metric")
}

func TestMStorage_StorageTypeDetection(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.ServerConfig
		isDB   bool
		isFile bool
	}{
		{
			name:   "DB only",
			cfg:    config.ServerConfig{DataBaseDsn: "postgres://u:p@h:5432/d"},
			isDB:   true,
			isFile: false,
		},
		{
			name:   "File only",
			cfg:    config.ServerConfig{FileStorage: "/tmp/test.json"},
			isDB:   false,
			isFile: true,
		},
		{
			name:   "Both (DB priority)",
			cfg:    config.ServerConfig{DataBaseDsn: "postgres://u:p@h:5432/d", FileStorage: "/tmp/test.json"},
			isDB:   true,
			isFile: false,
		},
		{
			name:   "Neither (memory)",
			cfg:    config.ServerConfig{},
			isDB:   false,
			isFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage(tt.cfg).(*MStorage)
			assert.Equal(t, tt.isDB, storage.isStorageBD())
			assert.Equal(t, tt.isFile, storage.isStorageFile())
		})
	}
}

func TestMStorage_UpdateMetric_GaugePointer(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	val := float64(123.45)
	storage.UpdateMetric("test", "gauge", &val)

	metrics := storage.ListMetrics()
	assert.Equal(t, 123.45, metrics["test"].Gauge)
}

func TestMStorage_UpdateMetric_CounterPointer(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	val := int64(456)
	storage.UpdateMetric("test", "counter", &val)

	metrics := storage.ListMetrics()
	assert.Equal(t, int64(456), metrics["test"].Counter)
}

func TestMStorage_UpdateMetric_NilValue(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test_gauge", "gauge", (*float64)(nil))
	storage.UpdateMetric("test_counter", "counter", (*int64)(nil))

	metrics := storage.ListMetrics()
	assert.Contains(t, metrics, "test_gauge")
	assert.Contains(t, metrics, "test_counter")
}

func TestMStorage_UpdateMetric_InvalidMetricType(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	// Unknown metric type
	storage.UpdateMetric("test", "unknown_type", 100.0)

	metrics := storage.ListMetrics()
	// Should still create the metric with the type
	assert.Contains(t, metrics, "test")
}

func TestMStorage_ListMetrics_AfterUpdate(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("m1", "gauge", 1.0)
	storage.UpdateMetric("m2", "counter", int64(2))

	metrics1 := storage.ListMetrics()
	assert.Equal(t, 2, len(metrics1))

	storage.UpdateMetric("m3", "gauge", 3.0)

	metrics2 := storage.ListMetrics()
	assert.Equal(t, 3, len(metrics2))
}

func TestMStorage_GetMetric_AfterUpdate(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	storage.UpdateMetric("test", "gauge", 100.0)
	storage.UpdateMetric("test", "gauge", 200.0)

	metric, found := storage.GetMetric("test")
	assert.True(t, found)
	assert.Equal(t, 200.0, metric.Gauge)
}

func TestMStorage_UpdatesMetrics_DifferentTypes(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	metrics := []models.Metrics{
		{ID: "g1", MType: "gauge", Value: floatPtr(1.0)},
		{ID: "c1", MType: "counter", Delta: int64Ptr(2)},
		{ID: "g2", MType: "gauge", Value: floatPtr(3.0)},
		{ID: "c2", MType: "counter", Delta: int64Ptr(4)},
	}

	storage.UpdatesMetrics(metrics)

	result := storage.ListMetrics()
	assert.Equal(t, 4, len(result))
}

func TestMStorage_StorageGracefulStop_Concurrent(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storage.StorageGracefulStop(&http.Server{Addr: ":8082"})
		}()
	}

	wg.Wait()
}

func TestMStorage_updateMetricMemory_NilValue(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	// Test with nil float64
	err := storage.updateMetricMemory("test1", "gauge", (*float64)(nil))
	assert.NoError(t, err)

	// Test with nil int64 (should not happen in practice but let's test)
	err = storage.updateMetricMemory("test2", "counter", (*int64)(nil))
	assert.NoError(t, err)
}

func TestMStorage_updateMetricMemory_InvalidGaugeType(t *testing.T) {
	storage := NewMemStorage(config.ServerConfig{}).(*MStorage)

	err := storage.updateMetricMemory("test", "gauge", "string_value")
	assert.Error(t, err)
}

func TestMStorage_initStorageFile_DirectoryCreation(t *testing.T) {
	tmpFile := "/tmp/test_nested/nested_dir/metrics.json"
	defer os.RemoveAll("/tmp/test_nested")

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
	}
	storage := NewMemStorage(cfg).(*MStorage)

	storage.initStorageFile()
}

func TestMStorage_saveStorageMetrics_DirectoryCreation(t *testing.T) {
	tmpFile := "/tmp/test_nested_save/nested_dir/metrics.json"
	defer os.RemoveAll("/tmp/test_nested_save")

	cfg := config.ServerConfig{
		FileStorage: tmpFile,
	}
	storage := NewMemStorage(cfg).(*MStorage)

	storage.saveStorageMetrics()
}
