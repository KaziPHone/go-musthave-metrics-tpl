package storage

import (
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

const (
	useBd = iota + 1
	useFileStorage
	useMemory
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
	dataBase      *dataBase
	useStorage int
}

type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

func NewMemStorage(cfg config.ServerConfig) IStorage {

	db := newDataBase(cfg.DataBaseDsn, cfg.MigratePath)
	db.initDataBase()

	storage := &MStorage{
		MetricTypes:   make(map[string]*MetricType),
		fileStorage:   cfg.FileStorage,
		storeInterval: cfg.StoreInterval,
		restore:       cfg.Restore,
		dataBase:      db,
	}

	storage.setStorage()

	storage.initStorageFile()

	return storage
}
