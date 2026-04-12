package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	httpDelivery "taskflow/delivery/http"
	"taskflow/internal/repository"
	"taskflow/internal/usecase"
	"taskflow/utils"
	postgres "taskflow/utils/database/postgres"
	"taskflow/utils/logger"
)

func main() {
	config := utils.LoadAndGetConfig()
	logger.Init(config)
	slog.Info("starting taskflow", "env", config.Env, "port", config.ServerPort)

	db, err := postgres.NewPool(config.DBConnectionStr)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repos := repository.New(db)
	uc := usecase.Init(config, repos)

	var wg sync.WaitGroup
	server := httpDelivery.NewRestDelivery(config, &uc, &wg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if server != nil {
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}

	wg.Wait()
	slog.Info("server stopped")
}
