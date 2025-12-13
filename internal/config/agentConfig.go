package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
}

func NewConfigAgent() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	err := cfg.initConfig()
	return cfg, err
}

func (cfg *AgentConfig) initConfig() error {

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "адрес HTTP-сервера")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "частота отправки метрик на сервер (по умолчанию 10 секунд)")
	flag.IntVar(&cfg.PollInterval, "p", 3, "частота опроса метрик из пакета runtime (по умолчанию 2 секунды)")
	flag.StringVar(&cfg.Key, "k", "", "Ключ")

	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		return errors.New("переданы неизвестные флаги")
	}

	if err := env.Parse(cfg); err != nil {
		return err
	}

	return nil
}
