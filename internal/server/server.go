package server

import (
	"context"
	"errors"
	"github.com/salivare/auth-server/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const DefaultShutdownTimeout = 10 * time.Second

func New(handler http.Handler, cfgServer config.ServerConfig) *http.Server {

	return &http.Server{
		Addr:         cfgServer.Addr,
		Handler:      handler,
		ReadTimeout:  cfgServer.ReadTimeout,
		WriteTimeout: cfgServer.WriteTimeout,
		IdleTimeout:  cfgServer.IdleTimeout,
	}
}

func Start(srv *http.Server) error {
	if srv == nil {
		return errors.New("nil server")
	}

	errCn := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCn <- err
			return
		}

		errCn <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s, start shutdown", sig)
	case err := <-errCn:
		if err == nil {
			log.Printf("server stopped without error")
			return nil
		}
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Printf("server shutdown complete")

	return nil
}
