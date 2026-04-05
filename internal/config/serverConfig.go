package config

import (
	"flag"

	"github.com/caarlos0/env"
)

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

// NewConfigServer создает новый конфиг сервера, считывая из флагов и окружения
func NewConfigServer() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	fs := flag.NewFlagSet("server-config", flag.ContinueOnError)

	fs.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	fs.StringVar(&cfg.FileStorage, "f", "metrics_storage.json", "Путь до файла, куда сохраняются текущие значения")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	fs.BoolVar(&cfg.Restore, "r", false, "Восстановление данных из файла, если он существует")
	fs.StringVar(&cfg.DataBaseDsn, "d", "", "Строка подключения к базе данных")
	fs.StringVar(&cfg.Key, "k", "", "Ключ")

	// Парсим с nil args - просто устанавливаем значения по умолчанию
	if err := fs.Parse(nil); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
