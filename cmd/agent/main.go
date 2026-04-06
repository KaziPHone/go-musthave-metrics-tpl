package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	done := make(chan struct{})

	// Запускаем агента в отдельной горутине, чтобы main мог ловить сигналы
	go func() {
		agentMetric.Start(stop)
		close(done)
	}()

	// Ловим сигналы и аккуратно останавливаем агента
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh
	log.Print("Signal received, stopping agent...")
	close(stop)
	// Ждём завершения агента
	<-done
	log.Print("Agent stopped")
}
