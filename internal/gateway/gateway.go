package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agentfence/agentfence/internal/api"
	"github.com/agentfence/agentfence/internal/config"
)

// Gateway is the top-level runtime for the HTTP gateway process.
type Gateway struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
}

// New constructs a Gateway with explicit configuration.
func New(cfg config.Config, logger *slog.Logger) *Gateway {
	handler := api.NewHandler(logger)

	return &Gateway{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:              cfg.HTTP.Address,
			Handler:           handler,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
		},
	}
}

// ListenAddr returns the configured bind address for the gateway process.
func (g *Gateway) ListenAddr() string {
	return g.cfg.HTTP.Address
}

// Run starts the HTTP server and shuts it down when the context is canceled.
func (g *Gateway) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	g.logger.Info("starting gateway",
		"environment", g.cfg.Environment,
		"address", g.server.Addr,
	)

	go func() {
		err := g.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve gateway: %w", err)
		}
		return nil
	case <-ctx.Done():
		g.logger.Info("gateway shutdown requested")
		if err := g.Shutdown(context.Background()); err != nil {
			return err
		}
		return <-errCh
	}
}

// Shutdown gracefully stops the HTTP server.
func (g *Gateway) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, g.cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := g.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown gateway: %w", err)
	}

	g.logger.Info("gateway shutdown complete")
	return nil
}
