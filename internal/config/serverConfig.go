// Package config предоставляет конфигурацию для сервера и агента.
//
// Конфигурация читается из флагов командной строки и переменных окружения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
//   - TrustedSubnet: доверенная подсеть в формате CIDR (например "192.168.0.0/24")
type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
	DataBaseDsn   string `env:"DATABASE_DSN"`
	MigratePath   string `env:"MIGRATE_PATH"`
	Key           string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
	GRPCAddress   string `env:"GRPC_ADDRESS"`
}

// AgentConfig конфигурация агента для сбора и отправки метрик.
//
// Поля:
//   - Host: адрес хоста, куда отправляются метрики
//   - ReportInterval: интервал отправки метрик в секундах
//   - PollInterval: интервал опроса метрик в секундах
//   - Key: ключ для SHA256 хэширования
//   - RateLimit: лимит одновременных запросов
type AgentConfig struct {
	Host           string `env:"AGENT_HOST"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	GRPCAddress    string `env:"GRPC_ADDRESS"`
}

// NewConfigServer создает новую конфигурацию сервера, считывая из флагов и окружения.
//
// Приоритет:
//  1. Флаги командной строки
//  2. Переменные окружения
//  3. Значения по умолчанию
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
	// Устанавливаем значения по умолчанию
	cfg := &ServerConfig{
		Host:          "localhost:8080",
		FileStorage:   "metrics_storage.json",
		StoreInterval: 300,
		Restore:       false,
		DataBaseDsn:   "",
		MigratePath:   "",
		Key:           "",
		CryptoKey:     "",
		AuditFile:     "",
		AuditURL:      "",
	}

	// 1) прочитать путь к файлу конфигурации из флага -c/-config или переменной CONFIG
	var cfgPath string
	cfgFs := flag.NewFlagSet("cfgfile", flag.ContinueOnError)
	cfgFs.SetOutput(io.Discard)
	cfgFs.StringVar(&cfgPath, "c", "", "path to config JSON file")
	cfgFs.StringVar(&cfgPath, "config", "", "path to config JSON file")
	// фильтруем тестовые флаги перед парсингом
	argsForCfg := os.Args[1:]
	filteredCfg := make([]string, 0, len(argsForCfg))
	for _, a := range argsForCfg {
		if strings.HasPrefix(a, "-test.") {
			continue
		}
		filteredCfg = append(filteredCfg, a)
	}
	_ = cfgFs.Parse(filteredCfg)
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG")
	}

	// 2) если файл конфигурации указан — прочитать и применить значения (они будут иметь меньший приоритет чем env/flags)
	if cfgPath != "" {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", cfgPath, err)
		}
		// временная структура с pointer-полями, чтобы отличать отсутствующие поля
		var fileCfg struct {
			Address       *string `json:"address"`
			Restore       *bool   `json:"restore"`
			StoreInterval *string `json:"store_interval"`
			StoreFile     *string `json:"store_file"`
			DatabaseDsn   *string `json:"database_dsn"`
			CryptoKey     *string `json:"crypto_key"`
			TrustedSubnet *string `json:"trusted_subnet"`
			GRPCAddress   *string `json:"grpc_address"`
		}
		if err := json.Unmarshal(data, &fileCfg); err != nil {
			return nil, fmt.Errorf("invalid json config %s: %w", cfgPath, err)
		}
		if fileCfg.Address != nil && *fileCfg.Address != "" {
			cfg.Host = *fileCfg.Address
		}
		if fileCfg.Restore != nil {
			cfg.Restore = *fileCfg.Restore
		}
		if fileCfg.StoreInterval != nil && *fileCfg.StoreInterval != "" {
			if d, err := time.ParseDuration(*fileCfg.StoreInterval); err == nil {
				cfg.StoreInterval = int(d.Seconds())
			}
		}
		if fileCfg.StoreFile != nil && *fileCfg.StoreFile != "" {
			cfg.FileStorage = *fileCfg.StoreFile
		}
		if fileCfg.DatabaseDsn != nil {
			cfg.DataBaseDsn = *fileCfg.DatabaseDsn
		}
		if fileCfg.CryptoKey != nil {
			cfg.CryptoKey = *fileCfg.CryptoKey
		}
		if fileCfg.TrustedSubnet != nil {
			cfg.TrustedSubnet = *fileCfg.TrustedSubnet
		}
		if fileCfg.GRPCAddress != nil {
			cfg.GRPCAddress = *fileCfg.GRPCAddress
		}
	}

	// 3) применяем переменные окружения (они имеют приоритет над файлом конфигурации)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// 4) наконец применяем флаги командной строки — они имеют самый высокий приоритет
	fs := flag.NewFlagSet("server-config", flag.ContinueOnError)
	// Не выводим usage в stdout/stderr (например, в тестах это ломало вывод примеров)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.Host, "a", cfg.Host, "адрес HTTP-сервера")
	fs.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "Путь до файла, куда сохраняются текущие значения")
	fs.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Восстановление данных из файла, если он существует")
	fs.StringVar(&cfg.DataBaseDsn, "d", cfg.DataBaseDsn, "Строка подключения к базе данных")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "Ключ")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Путь до PEM-файла приватного ключа или CRYPTO_KEY")
	fs.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Доверенная подсеть в формате CIDR (TRUSTED_SUBNET)")
	fs.StringVar(&cfg.GRPCAddress, "g", cfg.GRPCAddress, "Адрес gRPC сервера (GRPC_ADDRESS)")
	// парсим реальные args — удаляем специальные go test флаги, чтобы не ломать парсинг
	argsForFlags := os.Args[1:]
	filteredFlags := make([]string, 0, len(argsForFlags))
	for _, a := range argsForFlags {
		if strings.HasPrefix(a, "-test.") {
			continue
		}
		filteredFlags = append(filteredFlags, a)
	}
	_ = fs.Parse(filteredFlags)

	return cfg, nil
}

// NewConfigAgent создает новую конфигурацию агента, считывая из флагов и окружения.
//
// Приоритет:
//  1. Флаги командной строки
//  2. Переменные окружения
//  3. Значения по умолчанию
//
// Возвращает:
//   - *AgentConfig: инициализированная конфигурация
//   - error: ошибка при парсинге или nil
//
// Пример:
//
//	cfg, err := config.NewConfigAgent()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Agent will connect to %s\n", cfg.Host)
func NewConfigAgent() (*AgentConfig, error) {
	// значения по умолчанию
	cfg := &AgentConfig{
		Host:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   5,
		Key:            "",
		RateLimit:      5,
		CryptoKey:      "",
	}

	// 1) получить путь к файлу конфигурации из -c/-config или env CONFIG
	var cfgPath string
	cfgFs := flag.NewFlagSet("cfgfile", flag.ContinueOnError)
	cfgFs.SetOutput(io.Discard)
	cfgFs.StringVar(&cfgPath, "c", "", "path to config JSON file")
	cfgFs.StringVar(&cfgPath, "config", "", "path to config JSON file")
	argsForCfg := os.Args[1:]
	filteredCfg := make([]string, 0, len(argsForCfg))
	for _, a := range argsForCfg {
		if strings.HasPrefix(a, "-test.") {
			continue
		}
		filteredCfg = append(filteredCfg, a)
	}
	_ = cfgFs.Parse(filteredCfg)
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG")
	}

	// 2) прочитать файл (если указан) и применить значения с низким приоритетом
	if cfgPath != "" {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", cfgPath, err)
		}
		var fileCfg struct {
			Address        *string `json:"address"`
			ReportInterval *string `json:"report_interval"`
			PollInterval   *string `json:"poll_interval"`
			CryptoKey      *string `json:"crypto_key"`
			GRPCAddress    *string `json:"grpc_address"`
		}
		if err := json.Unmarshal(data, &fileCfg); err != nil {
			return nil, fmt.Errorf("invalid json config %s: %w", cfgPath, err)
		}
		if fileCfg.Address != nil && *fileCfg.Address != "" {
			cfg.Host = *fileCfg.Address
		}
		if fileCfg.ReportInterval != nil && *fileCfg.ReportInterval != "" {
			if d, err := time.ParseDuration(*fileCfg.ReportInterval); err == nil {
				cfg.ReportInterval = int(d.Seconds())
			}
		}
		if fileCfg.PollInterval != nil && *fileCfg.PollInterval != "" {
			if d, err := time.ParseDuration(*fileCfg.PollInterval); err == nil {
				cfg.PollInterval = int(d.Seconds())
			}
		}
		if fileCfg.CryptoKey != nil {
			cfg.CryptoKey = *fileCfg.CryptoKey
		}
		if fileCfg.GRPCAddress != nil {
			cfg.GRPCAddress = *fileCfg.GRPCAddress
		}
	}

	// 3) применяем env (они имеют приоритет над файлом)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// 4) применяем флаги (самый высокий приоритет)
	fs := flag.NewFlagSet("agent-config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.Host, "a", cfg.Host, "адрес хоста для отправки метрик")
	fs.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "интервал отправки метрик в секундах")
	fs.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "интервал опроса метрик в секундах")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "ключ для SHA256 хэширования")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "лимит одновременных запросов")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Путь до PEM-файла публичного ключа или CRYPTO_KEY")
	fs.StringVar(&cfg.GRPCAddress, "g", cfg.GRPCAddress, "Адрес gRPC сервера (GRPC_ADDRESS)")
	argsForFlags := os.Args[1:]
	filteredFlags := make([]string, 0, len(argsForFlags))
	for _, a := range argsForFlags {
		if strings.HasPrefix(a, "-test.") {
			continue
		}
		filteredFlags = append(filteredFlags, a)
	}
	_ = fs.Parse(filteredFlags)

	return cfg, nil
}
