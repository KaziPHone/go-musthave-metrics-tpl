package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type ServerConfig struct {
	Host string `env:"ADDRESS"`
}

func NewConfigServer() *ServerConfig {
	cfg := &ServerConfig{}

	env.Parse(cfg)

	if os.Getenv("ADDRESS") == "" {
		flag.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	}

	flag.Parse()

	return cfg
}
