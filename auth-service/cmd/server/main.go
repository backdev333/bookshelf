package main

import (
	"bookshelf/auth-service/internal/config"
	"bookshelf/auth-service/internal/handler"
	"bookshelf/auth-service/internal/repository"
	"bookshelf/auth-service/internal/service"
	"log"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.Load()
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to db error: %s", err.Error())
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	userServcie := service.NewUserService(repo, cfg.JWTSecret)
	h := handler.New(userServcie, cfg.JWTSecret)
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/health", h.Health)

	slog.Info("auth-service started", "port", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, nil)

}
