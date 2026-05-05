package main

import (
	"context"
	"errors"
	"frontdev333/bookshelf/internal/config"
	"frontdev333/bookshelf/internal/handler"
	"frontdev333/bookshelf/internal/repository"
	"frontdev333/bookshelf/internal/service"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const stdTimeout = 30 * time.Second

func main() {
	cfg := config.Load()

	port := cfg.Port

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to db error: %s", err.Error())
	}
	defer db.Close()

	slog.Info("connected to database")
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Minute)

	repos := repository.New(db)
	services := service.New(repos, cfg.JWTSecret)
	handlers := handler.New(services, cfg.JWTSecret)

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(stdTimeout))

	mux.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health)
		r.Get("/ready", handlers.Ready)
		r.Post("/auth/register", handlers.Register)
		r.Post("/auth/login", handlers.Login)
		r.Get("/books", handlers.ListBook)
		r.Get("/books/{bookId}", handlers.GetBook)
		r.Get("/books/{bookId}/reviews", handlers.ListBookReviews)
		r.Get("/reviews/{reviewId}", handlers.GetReview)

		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)

			r.Get("/users/me", handlers.GetCurrentUser)
			r.Put("/users/me", handlers.UpdateCurrentUser)
			r.Post("/books", handlers.CreateBook)
			r.Put("/books/{bookId}", handlers.UpdateBook)
			r.Delete("/books/{bookId}", handlers.DeleteBook)
			r.Post("/books/{bookId}/reviews", handlers.CreateReview)
			r.Put("/reviews/{reviewId}", handlers.UpdateReview)
			r.Delete("/reviews/{reviewId}", handlers.DeleteReview)
		})
	})

	slog.Info("Server starting", "port", port)

	finish := make(chan os.Signal, 1)
	server := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadTimeout:       stdTimeout,
		ReadHeaderTimeout: stdTimeout,
		WriteTimeout:      stdTimeout,
		IdleTimeout:       stdTimeout,
	}

	go func() {
		if err = server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	signal.Notify(finish, os.Interrupt, syscall.SIGTERM)
	<-finish

	ctx, cancel := context.WithTimeout(context.Background(), stdTimeout)
	defer cancel()
	server.Shutdown(ctx)
}
