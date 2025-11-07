package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host string `env:"ADDRESS"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	StoreInterval int `env:"STORE_INTERVAL"`
	Restore bool `env:"RESTORE"`
	DataBaseDsn string `env:"DATABASE_DSN"`
	MigratePath string `env:"MIGRATE_PATH"`
}

func NewConfigServer() *ServerConfig {
	cfg := &ServerConfig{}

	env.Parse(cfg)

	if os.Getenv("ADDRESS") == "" {
		flag.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	}

	if os.Getenv("FILE_STORAGE_PATH") == "" {
		flag.StringVar(&cfg.FileStorage, "f", "metrics_storage.json", "Путь до файла, куда сохраняются текущие значения")
	}
	
	if os.Getenv("STORE_INTERVAL") == "" {
		flag.IntVar(&cfg.StoreInterval, "i", 300, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	}

	if os.Getenv("RESTORE") == "" {
		flag.BoolVar(&cfg.Restore, "r", false, "Восстановление данных из файла, если он существует")
	}

	if os.Getenv("DATABASE_DSN") == "" {
		flag.StringVar(&cfg.DataBaseDsn, "d", "", "Строка подключения к базе данных")
	}


	flag.Parse()

	return cfg
}