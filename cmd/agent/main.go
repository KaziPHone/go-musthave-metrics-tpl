package main

import (
	"fmt"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
)

func main() {
	fmt.Println("START AGENT")
	agent := agent.NewAgent()
	stop := make(chan struct{})
	agent.Start(stop)
}
