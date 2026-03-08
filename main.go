package main

import (
	"crud-in-memory/api"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	slog.Info("system running")
	if err := run(); err != nil {
		slog.Error("failed to execute server", "error", err)
		os.Exit(1)
	}
	slog.Info("all system offline")
}

func run() error {
	db := api.ApplicationDB{
		Data: make(map[api.Id]api.User),
	}

	handler := api.NewHandler(db)

	s := http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
		Addr:         ":8080",
		Handler:      handler,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
