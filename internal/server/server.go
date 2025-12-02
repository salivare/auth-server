package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Config contains server parameters
type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Handler         http.Handler
}

// Server wraps over http.Server
type Server struct {
	cfg *Config
	srv *http.Server
}

// New creates a configuration-based Server
func New(cfg *Config, handler http.Handler) *Server {
	if cfg == nil {
		cfg = &Config{}
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}

	httpSrv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		cfg: cfg,
		srv: httpSrv,
	}
}

// Serve starts the server and blocks until it is finished
// Behavior: runs ListenAndServe in the server, waiting for ctx. Done() or ListenAndServe error
// When the context is canceled, graceful shutdown is executed with a cfg.ShutdownTimeout timer
func (s *Server) Serve(ctx context.Context) error {
	errCh := make(chan error, 1)

	// Running the server in background
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		// Context cleared - initiate graceful shutdown
		shutdownCtx := ctx
		if s.cfg.ShutdownTimeout > 0 {
			var cancel context.CancelFunc
			shutdownCtx, cancel = context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
			defer cancel()
		}
		return s.srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		// Server failed
		return err
	}
}
