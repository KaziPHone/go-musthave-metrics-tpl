// Package config предоставляет конфигурацию сервера.
//
// Конфигурация читается из флагов командной строки и переменных окружения.
package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// ServerConfig конфигурация HTTP-сервера для метрик.
//
// Поля:
//   - Host: адрес HTTP-сервера (по умолчанию "localhost:8080")
//   - FileStorage: путь к файлу для сохранения метрик
//   - StoreInterval: интервал сохранения в секундах (по умолчанию 300)
//   - Restore: флаг восстановления из файла (по умолчанию false)
//   - DataBaseDsn: строка подключения к базе данных
//   - MigratePath: путь к миграциям базы данных
//   - Key: ключ для SHA256 хэширования
//   - AuditFile: путь к файлу аудита
//   - AuditURL: URL для HTTP уведомлений аудита
type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
	DataBaseDsn   string `env:"DATABASE_DSN"`
	MigratePath   string `env:"MIGRATE_PATH"`
	Key           string `env:"KEY"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
}

// NewConfigServer создает новую конфигурацию сервера, считывая из флагов и окружения.
//
// Приоритет:
//   1. Флаги командной строки
//   2. Переменные окружения
//   3. Значения по умолчанию
//
// Возвращает:
//   - *ServerConfig: инициализированная конфигурация
//   - error: ошибка при парсинге или nil
//
// Пример:
//
//	cfg, err := config.NewConfigServer()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Server will run on %s\n", cfg.Host)
func NewConfigServer() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	fs := flag.NewFlagSet("server-config", flag.ContinueOnError)

	fs.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	fs.StringVar(&cfg.FileStorage, "f", "metrics_storage.json", "Путь до файла, куда сохраняются текущие значения")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	fs.BoolVar(&cfg.Restore, "r", false, "Восстановление данных из файла, если он существует")
	fs.StringVar(&cfg.DataBaseDsn, "d", "", "Строка подключения к базе данных")
	fs.StringVar(&cfg.Key, "k", "", "Ключ")

	// Parse with nil args - just set defaults
	if err := fs.Parse(nil); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
