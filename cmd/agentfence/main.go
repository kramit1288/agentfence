package main

import (
	"log"

	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/gateway"
)

func main() {
	cfg := config.Default()
	app := gateway.New(cfg)

	log.Printf("agentfence gateway starting on %s", app.ListenAddr())
}
