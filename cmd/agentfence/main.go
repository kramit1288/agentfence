package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentfence/agentfence/internal/approval"
	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/gateway"
	"github.com/agentfence/agentfence/internal/mcp/transport"
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
	approvalService := approval.NewService(approval.NewFileRepository(approvalStorePath()))
	options := []gateway.Option{gateway.WithApprovalManager(approvalService)}
	if upstreamURL := os.Getenv("AGENTFENCE_UPSTREAM_URL"); upstreamURL != "" {
		forwarder, err := transport.NewHTTPForwarder(transport.Target{Address: upstreamURL}, &http.Client{Timeout: upstreamTimeout()})
		if err != nil {
			return fmt.Errorf("configure upstream forwarder: %w", err)
		}
		options = append(options, gateway.WithForwarder(forwarder))
	}
	app := gateway.New(cfg, logger, options...)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return app.Run(ctx)
}

func approvalStorePath() string {
	if path := os.Getenv("AGENTFENCE_APPROVAL_STORE"); path != "" {
		return path
	}
	return "data/approvals.json"
}

func upstreamTimeout() time.Duration {
	if raw := os.Getenv("AGENTFENCE_UPSTREAM_TIMEOUT"); raw != "" {
		if value, err := time.ParseDuration(raw); err == nil {
			return value
		}
	}
	return 10 * time.Second
}
