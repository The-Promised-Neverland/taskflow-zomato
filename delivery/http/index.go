package http

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"taskflow/delivery/http/routes"
	"taskflow/internal/usecase"
	"taskflow/utils"
)

func NewRestDelivery(config *utils.Config, uc *usecase.UseCases, wg *sync.WaitGroup) *http.Server {
	router := routes.NewRouter(uc)

	server := &http.Server{
		Addr:              ":" + config.ServerPort,
		Handler:           router,
		ReadTimeout:       config.ServerReadTimeout,
		ReadHeaderTimeout: config.ServerHeaderTimeout,
		WriteTimeout:      config.ServerWriteTimeout,
		IdleTimeout:       config.ServerIdleTimeout,
	}

	listener, err := net.Listen("tcp", ":"+config.ServerPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting API server", "addr", server.Addr)
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	return server
}
