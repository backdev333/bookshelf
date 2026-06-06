package main

import (
	"bookshelf/auth-service/internal/config"
	"bookshelf/auth-service/internal/handler"
	"bookshelf/auth-service/internal/repository"
	"bookshelf/auth-service/internal/service"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.Load()
	db := sqlx.Connect("postgres", cfg.DatabaseURL)
	repo := repository.New(db)
	userServcie := service.New(repo, cfg.JWTSecret)
	h := handler.New(userServcie, cfg.JWTSecret)
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/health", h.Health)

	slog.Info("auth-service started", "port", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, nil)

}
