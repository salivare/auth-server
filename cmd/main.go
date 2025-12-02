package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/salivare/auth-server/internal/config"
	"github.com/salivare/auth-server/internal/handler/auth"
	"github.com/salivare/auth-server/internal/server"
	"github.com/salivare/auth-server/internal/service"
	"github.com/salivare/auth-server/internal/storage/sqlite"
	"github.com/salivare/auth-server/internal/token"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	cfg, confErr := config.LoadConfig()
	if confErr != nil {
		log.Fatal(confErr)
	}

	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "file:auth.db?_foreign_keys=1"
	}
	db, err := sqlite.Open(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		db.Close()
	}()

	v := validator.New()
	mux := http.NewServeMux()

	userStore := sqlite.NewUserRepo(db)
	tokenStore := sqlite.NewTokenStore(db)

	secret := []byte(cfg.Token.Secret)
	accessTTL := time.Minute * 15
	refreshTTL := time.Hour * 24
	tokenManager := token.NewManager(secret, accessTTL)

	svc := service.NewAuthService(userStore, tokenStore, tokenManager, refreshTTL)

	h := auth.NewHandler(svc)

	serverConfig := &server.Config{
		Addr:            cfg.Server.Addr,
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		IdleTimeout:     cfg.Server.IdleTimeout,
		ShutdownTimeout: 10 * time.Second,
	}

	srv := server.New(serverConfig, mux)
	auth.RegisterRoutes(mux, h, v)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Serve(ctx); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}

	log.Println("server stopped gracefully")
}
