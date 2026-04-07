// Package storage предоставляет интерфейс и реализации хранилища метрик.
//
// Реализует три режима хранения:
//   - Memory: хранение в памяти (по умолчанию)
//   - File: хранение в JSON файле
//   - Database: хранение в PostgreSQL
//
// Поддерживает конкурентный доступ с помощью sync.RWMutex.
package storage

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/rs/zerolog/log"
)

const (
	// UseBd режим хранения в базе данных.
	UseBd = iota + 1
	// UseFileStorage режим хранения в файле.
	UseFileStorage
	// UseMemory режим хранения в памяти.
	UseMemory
)

// IStorage интерфейс хранилища метрик.
//
// Определяет методы для:
//   - Обновления метрик (UpdateMetric, UpdatesMetrics)
//   - Получения метрик (ListMetrics, GetMetric)
//   - Управления жизненным циклом (StorageGracefulStop)
//   - Проверки подключения (IsConnectedDB)
type IStorage interface {
	// UpdateMetric обновляет одну метрику по имени.
	//
	// Параметры:
	//   - metricName: имя метрики
	//   - typeMetric: тип метрики ("gauge" или "counter")
	//   - value: значение метрики (float64 для gauge, int64 для counter)
	UpdateMetric(metricName, typeMetric string, value interface{})

	// UpdatesMetrics обновляет несколько метрик за раз.
	//
	// Параметры:
	//   - metrics: срез метрик для обновления
	UpdatesMetrics([]models.Metrics)

	// ListMetrics возвращает все метрики.
	//
	// Возвращает:
	//   - map[string]*MetricType: карта метрик по имени
	ListMetrics() map[string]*MetricType

	// GetMetric возвращает метрику по имени.
	//
	// Параметры:
	//   - metricName: имя метрики
	//
	// Возвращает:
	//   - *MetricType: метрика или nil если не найдена
	//   - bool: true если метрика найдена
	GetMetric(metricName string) (*MetricType, bool)

	// StorageGracefulStop останавливает хранилище с сохранением данных.
	//
	// Параметры:
	//   - server: сервер HTTP для корректной остановки
	StorageGracefulStop(server *http.Server)

	// IsConnectedDB проверяет подключение к базе данных.
	//
	// Возвращает:
	//   - bool: true если подключено к БД
	IsConnectedDB() bool
}

// MStorage реализация IStorage с поддержкой разных типов хранилищ.
//
// Поддерживает три режима хранения (memory, file, database) и автоматически
// переключается между ними в зависимости от конфигурации.
type MStorage struct {
	mu            sync.RWMutex
	MetricTypes   map[string]*MetricType
	fileStorage   string `env:"FILE_STORAGE_PATH"`
	storeInterval int    `env:"STORE_INTERVAL"`
	restore       bool   `env:"RESTORE"`
	dataBase      *dataBase
	useStorage    int
}

// MetricType представляет тип метрики с текущим значением.
//
// Одновременно может хранить либо Gauge (float64), либо Counter (int64).
type MetricType struct {
	Mtype   string
	Gauge   float64
	Counter int64
}

// NewMemStorage создает новое хранилище метрик на основе конфигурации.
//
// Параметры:
//   - cfg: конфигурация сервера (см. config.ServerConfig)
//
// Возвращает:
//   - IStorage: реализация интерфейса хранилища
//
// Приоритет выбора хранилища:
//  1. PostgreSQL (если DataBaseDsn задан)
//  2. File (если FileStorage задан)
//  3. Memory (по умолчанию)
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

// isStorageBD проверяет, используется ли база данных для хранения.
//
// Возвращает:
//   - bool: true если используется БД
func (m *MStorage) isStorageBD() bool {
	return m.useStorage == UseBd
}

// isStorageFile проверяет, используется ли файл для хранения.
//
// Возвращает:
//   - bool: true если используется файл
func (m *MStorage) isStorageFile() bool {
	return m.useStorage == UseFileStorage
}

// setStorage устанавливает тип хранилища на основе конфигурации.
//
// Приоритет:
//  1. PostgreSQL (если DataBaseDsn задан)
//  2. File (если FileStorage задан)
//  3. Memory (по умолчанию)
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

// IsConnectedDB проверяет подключение к базе данных.
//
// Возвращает:
//   - bool: true если успешно подключено к БД
func (m *MStorage) IsConnectedDB() bool {
	return m.dataBase.isConnected
}

// StorageGracefulStop останавливает сервер и хранилище с сохранением данных.
//
// Параметры:
//   - server: сервер HTTP для корректной остановки
//
// Метод слушает сигналы SIGTERM и SIGINT, при получении которых:
//   - Сохраняет данные в файл или закрывает БД
//   - Логирует событие
//   - Завершает работу процесса
func (m *MStorage) StorageGracefulStop(server *http.Server) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		sig := <-stopChan

		log.Info().Msgf("Signal %s received, initiating graceful shutdown...", sig.String())

		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Err(err).Msg("error during server shutdown")
			}
		}

		switch {
		case m.isStorageBD():
			m.dataBase.CloseDataBase()
		case m.isStorageFile():
			m.saveStorageMetrics()
		default:
		}

		log.Info().Msg("Storage saved, shutdown complete")
	}()
}
