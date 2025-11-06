package main

import (
	"log"
	"os"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

func main() {
	cfg, err := config.NewConfigAgent()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	agentMetric := agent.NewAgent(*cfg)
	stop := make(chan struct{})
	agentMetric.Start(stop)
}
