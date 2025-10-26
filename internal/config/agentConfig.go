package config

import (
	"errors"
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func NewConfigAgent() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	err := cfg.initConfig()
	return cfg, err
}

func (cfg *AgentConfig) initConfig() error {

	err := env.Parse(cfg)

	if os.Getenv("ADDRESS") == "" {
		flag.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	}
	if os.Getenv("REPORT_INTERVAL") == "" {
		flag.IntVar(&cfg.ReportInterval, "r", 10, "частота отправки метрик на сервер (по умолчанию 10 секунд)")
	}
	if os.Getenv("POLL_INTERVAL") == "" {
		flag.IntVar(&cfg.PollInterval, "p", 3, "частота опроса метрик из пакета runtime (по умолчанию 2 секунды)")
	}

	args := flag.Args()
	if len(args) > 0 {
		return errors.New("переданы неизвестные флаги")
	}

	flag.Parse()

	return err
}
