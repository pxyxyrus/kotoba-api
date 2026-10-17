// Command api runs Kotoba's dedicated Go backend.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"kotoba/api/internal/database"
	"kotoba/api/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	gin.SetMode(gin.ReleaseMode)
	addr := os.Getenv("KOTOBA_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	var ready func(context.Context) error
	if uri := os.Getenv("KOTOBA_MONGO_URI"); uri != "" {
		connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		client, err := database.Open(connectCtx, uri)
		cancel()
		if err != nil {
			return err
		}
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := client.Disconnect(cleanup); err != nil {
				logger.Warn("database disconnect failed")
			}
		}()
		ready = func(ctx context.Context) error { return client.Ping(ctx, nil) }
	}
	server := &http.Server{
		Addr: addr, Handler: httpapi.NewRouterWithReadiness(logger, ready),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	stopped := make(chan error, 1)
	go func() { stopped <- server.ListenAndServe() }()
	logger.Info("server starting", "address", addr)
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		stop() // A second interrupt can terminate an unresponsive shutdown.
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return errors.Join(err, server.Close())
	}
	if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped gracefully")
	return nil
}
