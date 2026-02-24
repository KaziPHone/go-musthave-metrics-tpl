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
	AuditUrl      string `env:"AUDIT_URL"`
}

func NewConfigServer() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	flag.StringVar(&cfg.FileStorage, "f", "metrics_storage.json", "Путь до файла, куда сохраняются текущие значения")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.BoolVar(&cfg.Restore, "r", false, "Восстановление данных из файла, если он существует")
	flag.StringVar(&cfg.DataBaseDsn, "d", "", "Строка подключения к базе данных")
	flag.StringVar(&cfg.Key, "k", "", "Ключ")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
