package main

import (
	"encoding/json"
	"fmt"
	"frontdev333/bookshelf/internal/config"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	port := cfg.Port

	http.HandleFunc("/health", healthHandler)

	slog.Info("Server starting", "port", port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
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
