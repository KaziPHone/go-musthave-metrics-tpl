package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"syscall"
	"time"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type dataBase struct {
	dataBaseDsn   string
	isConnected   bool
	db            *sql.DB
	migratePath   string
	maxRetries    int             // максимальное кол-во попыток
	retryInterval []time.Duration // массив задержек для попыток
}

func newDataBase(dataBaseDsn, migratePath string) *dataBase {
	return &dataBase{
		dataBaseDsn:   dataBaseDsn,
		migratePath:   migratePath,
		maxRetries:    3,                                                              // Три дополнительных попытки
		retryInterval: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}, // Задержки
	}
}

// IsConnectedDB true если подключена база данных
func (d *dataBase) IsConnectedDB() bool {
	return d.isConnected
}

// initDataBase инициализация подключения к базе данных
func (d *dataBase) initDataBase() {
	if d.dataBaseDsn == "" {
		return
	}

	// Открываем базовое соединение с несколькими попытками
	db, err := d.openDBWithRetry()
	if err != nil {
		log.Printf("Unable to establish a connection with the database after retries: %v\n", err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Print("Database not connected...")
		return
	}

	d.isConnected = true
	d.db = db
	log.Print("Database connected")

	d.migrateDB()
}

func (d *dataBase) openDBWithRetry() (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i <= d.maxRetries; i++ {
		db, err = sql.Open("pgx", d.dataBaseDsn)
		if err == nil {
			break
		}

		if !d.retryableError(err) {
			return nil, fmt.Errorf("non-retryable error occurred during DB initialization: %w", err)
		}

		if len(d.retryInterval) > i {
			waitTime := d.retryInterval[i]
			log.Printf("Initial connection attempt failed, retrying in %v...\n", waitTime)
			time.Sleep(waitTime)
		} else {
			break
		}
	}

	return db, err
}

// migrateDB миграция базы данных
func (d *dataBase) migrateDB() {

	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		log.Printf("error creating migration driver: %v\n", err)
		return
	}

	d.setMigratePath()
	migrator, err := migrate.NewWithDatabaseInstance(
		d.migratePath,
		"postgres",
		driver,
	)

	if err != nil {
		log.Printf("failed to create migrator instance: %v\n", err)
		return
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("migration failed: %v\n", err)
		return
	}

	log.Print("migration success")
}

// setMigratePath установка пути миграции
func (d *dataBase) setMigratePath() {
	if d.migratePath == "" {
		d.migratePath = "file://./migrations"
	}
}

// insertMetric вставка метрики в базу данных
func (d *dataBase) insertMetric(metricName, typeMetric string, val interface{}) {

	var value *float64
	var delta *int64

	if typeMetric == models.Gauge {
		value = val.(*float64)
	} else {
		delta = val.(*int64)
	}

	query := `
        INSERT INTO metrics (id_metric, type_metric, value_metric, delta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id_metric)
		DO UPDATE SET
    	value_metric = EXCLUDED.value_metric,
    	delta = CASE WHEN EXCLUDED.delta IS NOT NULL THEN metrics.delta + EXCLUDED.delta ELSE metrics.delta END;`

	_, err := d.executeWithRetry(context.Background(), query,
		metricName, typeMetric, value, delta)
	if err != nil {
		log.Printf("insert/update failed: %v\n", err)
	}

}

func (d *dataBase) getMetric(metricName string) (*MetricType, bool) {

	metric := models.Metrics{}

	row := d.db.QueryRow("SELECT * FROM metrics WHERE id_metric = $1", metricName)
	err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	m := MetricType{
		Mtype: metric.MType,
	}
	if metric.MType == models.Counter {
		m.Counter = *metric.Delta
	} else {
		m.Gauge = *metric.Value
	}

	return &m, true
}

func (d *dataBase) getMetrics() map[string]*MetricType {

	metrics := make(map[string]*MetricType)
	rows, err := d.db.Query("SELECT * FROM metrics")
	if err != nil {
		log.Print(err)
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics

		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
			log.Print(err)
		}

		m := MetricType{
			Mtype: metric.MType,
		}
		if metric.MType == models.Gauge {
			m.Gauge = *metric.Value
		} else {
			m.Counter = *metric.Delta
		}
		metrics[metric.ID] = &m

	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
	}

	return metrics
}

// CloseDataBase закрытие базы данных
func (d *dataBase) CloseDataBase() {
	d.db.Close()
}

// retryableError определяет, является ли ошибка временной и подлежит повторению
func (d *dataBase) retryableError(err error) bool {
	// Сначала проверяем ошибку PostgreSQL
	pgErr, isPgError := err.(*pgconn.PgError)
	if isPgError {
		switch pgErr.Code {
		case pgerrcode.ConnectionException:
			return true
		default:
			return false
		}
	}

	// Теперь проверяем ошибки OS/Sockets
	netErr, isNetError := err.(interface {
		Timeout() bool
		Temporary() bool
	})
	if isNetError && netErr.Temporary() {
		return true
	}

	// Отдельно проверяем конкретную ошибку ECONNREFUSED
	opErr, isOpError := err.(*net.OpError)
	if isOpError && opErr.Err == syscall.ECONNREFUSED {
		return true
	}

	return false
}

// executeWithRetry функция для выполнения SQL-запросов с возможностью повторений
func (d *dataBase) executeWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	for i := 0; i <= d.maxRetries; i++ {
		result, err = d.db.ExecContext(ctx, query, args...)
		if err == nil {
			break
		}
		fmt.Println("retry", i)

		if !d.retryableError(err) {
			return nil, fmt.Errorf("non-retryable error occurred: %w", err)
		}

		// Проверяем наличие оставшегося периода ожиданий
		if len(d.retryInterval) > i {
			waitTime := d.retryInterval[i]
			log.Printf("Connection exception detected, retrying in %v...", waitTime)
			time.Sleep(waitTime)
		} else {
			break
		}
	}

	return result, err
}
