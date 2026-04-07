package main

import (
	"log"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/buildinfo"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
)

func main() {
	log.Printf("Build version: %s", buildinfo.GetVersion())
	log.Printf("Build date: %s", buildinfo.GetDate())
	log.Printf("Build commit: %s", buildinfo.GetCommit())

	cfg, err := config.NewConfigAgent()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	agentMetric := agent.NewAgent(*cfg)
	stop := make(chan struct{})
	agentMetric.Start(stop)
}
