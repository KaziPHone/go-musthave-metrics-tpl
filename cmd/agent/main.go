package main

import (
	"log"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

func main() {
	cfg, err := config.NewConfigAgent()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	agentMetric := agent.NewAgent(*cfg)
	stop := make(chan struct{})
	agentMetric.Start(stop)
}
