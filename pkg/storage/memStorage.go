package storage

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/rs/zerolog/log"
)

const (
	UseBd = iota + 1
	UseFileStorage
	UseMemory
)

type IStorage interface {
	UpdateMetric(metricName, typeMetric string, value interface{})
	UpdatesMetrics([]models.Metrics)
	ListMetrics() map[string]*MetricType
	GetMetric(metricName string) (*MetricType, bool)
	StorageGracefulStop(server *http.Server)
	IsConnectedDB() bool
}

type MStorage struct {
	mu            sync.RWMutex
	MetricTypes   map[string]*MetricType
	fileStorage   string `env:"FILE_STORAGE_PATH"`
	storeInterval int    `env:"STORE_INTERVAL"`
	restore       bool   `env:"RESTORE"`
	dataBase      *dataBase
	useStorage    int
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

// isStorageBD проверка на использование БД
func (m *MStorage) isStorageBD() bool {
	return m.useStorage == UseBd
}

// isStorageFile проверка на использование файла
func (m *MStorage) isStorageFile() bool {
	return m.useStorage == UseFileStorage
}

// setStorage устанавливает тип хранилища
func (m *MStorage) setStorage() {
	switch {
	case m.dataBase.dataBaseDsn != "":
		m.useStorage = UseBd
	case m.fileStorage != "":
		m.useStorage = UseFileStorage
	default:
		m.useStorage = UseMemory
	}
}

// IsConnectedDB проверка на подключение к БД
func (m *MStorage) IsConnectedDB() bool {
	return m.dataBase.isConnected
}

// StorageGracefulStop остановка сервера и закрытие хранилища
func (m *MStorage) StorageGracefulStop(server *http.Server) {

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-stopChan

		switch {
		case m.isStorageBD():
			m.dataBase.CloseDataBase()
		case m.isStorageFile():
			m.saveStorageMetrics()
		default:
		}

		log.Print("Signal received, initiating close storage and graceful shutdown ...")
		os.Exit(0)
	}()
}
