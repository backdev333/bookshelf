package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func main() {
	port := ":8080"

	http.HandleFunc("/health", healthHandler)

	slog.Info("Server starting", "port", port)
	http.ListenAndServe(port, nil)
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

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
