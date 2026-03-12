package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/gateway"
	"github.com/agentfence/agentfence/internal/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string

	flags := flag.NewFlagSet("agentfence", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&configPath, "config", "", "path to a JSON config file")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := telemetry.NewLogger(cfg.Log)
	app := gateway.New(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.Run(ctx)
}
