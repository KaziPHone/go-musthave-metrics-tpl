package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
)

func main() {

	host := flag.String("a", "localhost:8080", "адрес HTTP-сервера")
	reportInterval := flag.Int("r", 10, "частота отправки метрик на сервер (по умолчанию 10 секунд)")
	pollInterval := flag.Int("p", 3, "частота опроса метрик из пакета runtime (по умолчанию 2 секунды)")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: переданы неизвестные флаги.")
		os.Exit(1)
	}

	fmt.Println("START AGENT")
	agent := agent.NewAgent(*reportInterval, *pollInterval, *host)
	stop := make(chan struct{})
	agent.Start(stop)
}
