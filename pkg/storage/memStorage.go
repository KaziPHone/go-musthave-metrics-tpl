package storage

import (
	"database/sql"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

type IStorage interface {
	UpdateMetric(metricName, typeMetric string, value interface{}) error
	ListMetrics() map[string]*MetricType
	GetMetric(metricName string) (*MetricType, bool)
	GracefulStop(server *http.Server)
	IsConnectedDB() bool
}

type MStorage struct {
	MetricTypes   map[string]*MetricType
	fileStorage   string `env:"FILE_STORAGE_PATH"`
	storeInterval int    `env:"STORE_INTERVAL"`
	restore       bool   `env:"RESTORE"`
	dataBase      dataBase
}

type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

func NewMemStorage(cfg config.ServerConfig) IStorage {
	storage := &MStorage{
		MetricTypes:   make(map[string]*MetricType),
		fileStorage:   cfg.FileStorage,
		storeInterval: cfg.StoreInterval,
		restore:       cfg.Restore,
		dataBase:      dataBase{dataBaseDsn: cfg.DataBaseDsn, db: &sql.DB{}},
	}

	storage.initStorageFile()
	storage.initDataBase()
	return storage
}
