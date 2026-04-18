package main

import (
	"encoding/json"
	"frontdev333/bookshelf/internal/config"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.RequestID)

	mux.Get("/health", healthHandler)

	cfg := config.Load()

	port := cfg.Port

	slog.Info("Server starting", "port", port)
	http.ListenAndServe(":"+cfg.Port, mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not supported. use GET."))
		return
	}

	bytes, err := json.MarshalIndent(struct {
		Status string `json:"status"`
	}{"ok"}, "", "	")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("json marshal", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
