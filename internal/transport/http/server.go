package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(ctx context.Context, port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done(): // Если получили сигнал на остановку
		return s.Shutdown(context.Background())
	case err := <-serverErr:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}
