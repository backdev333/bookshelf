package main

import (
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/handler"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	services := service.NewBookService(repos)
	bookHandler := handler.NewBookHandler(services)

	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(stdTimeout))

	mux.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", bookHandler.Health)
		r.Get("/ready", bookHandler.Ready)
		r.Get("/books", bookHandler.List)
		r.Get("/books/{id}", bookHandler.Get)
		r.Get("/books/{id}/reviews", bookHandler.ListReviews)
		r.Get("/reviews/{reviewId}", bookHandler.GetReview)

		r.Group(func(r chi.Router) {
			r.Use(bookHandler.AuthMiddleware)

			r.Get("/users/me", bookHandler.GetCurrentUser)
			r.Put("/users/me", bookHandler.UpdateCurrentUser)
			r.Post("/books", bookHandler.Create)
			r.Put("/books/{id}", bookHandler.Update)
			r.Delete("/books/{id}", bookHandler.Delete)
			r.Post("/books/{id}/reviews", bookHandler.CreateReview)
			r.Put("/reviews/{reviewId}", bookHandler.UpdateReview)
			r.Delete("/reviews/{reviewId}", bookHandler.DeleteReview)
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
