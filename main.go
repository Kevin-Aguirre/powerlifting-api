package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kevin-Aguirre/powerlifting-api/api"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	port := getEnv("PORT", "8080")
	repoURL := getEnv("REPO_URL", data.DefaultRepoURL)
	repoPath := getEnv("REPO_PATH", data.DefaultRepoPath)
	dataPath := getEnv("DATA_PATH", data.DefaultDataPath)

	refreshStr := getEnv("REFRESH_INTERVAL", "1h")
	refreshInterval, err := time.ParseDuration(refreshStr)
	if err != nil {
		slog.Error("invalid REFRESH_INTERVAL", "value", refreshStr, "error", err)
		os.Exit(1)
	}

	ds := data.NewDataStore()
	if err := ds.Init(repoPath, repoURL, dataPath, refreshInterval); err != nil {
		slog.Error("failed to initialize data", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: api.NewRouter(ds),
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server stopped")
}
