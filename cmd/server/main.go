package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ttwms/internal/config"
	"ttwms/internal/httpapp"
	"ttwms/internal/warehouse"
	"ttwms/migrations"
)

func main() {
	envFile := flag.String("env-file", "", "Read defaults from an explicit KEY=value file")
	migrateOnly := flag.Bool("migrate", false, "Apply database migrations and exit")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var cfg config.Config
	var err error
	if *envFile != "" {
		cfg, err = config.LoadFile(*envFile)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	if *migrateOnly && cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required for migrations")
		os.Exit(1)
	}
	options := httpapp.Options{Environment: string(cfg.Environment)}
	if cfg.DatabaseURL != "" {
		db, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			logger.Error("invalid PostgreSQL connection settings")
			os.Exit(1)
		}
		defer db.Close()
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err = db.PingContext(ctx)
		if err == nil {
			if *migrateOnly || cfg.Environment == config.Development || cfg.Environment == config.Test {
				err = migrations.Apply(ctx, db)
			} else {
				err = migrations.Check(ctx, db)
			}
		}
		cancel()
		if err != nil {
			// Connection errors may embed credentials or connection strings.
			logger.Error("PostgreSQL startup failed; check connectivity and run -migrate if schema updates are pending")
			os.Exit(1)
		}
		options.Warehouse = &warehouse.Postgres{DB: db}
		logger.Info("PostgreSQL ready; schema is current")
	}
	if *migrateOnly {
		logger.Info("database migrations complete")
		return
	}

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           httpapp.NewWithOptions(logger, options),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server listening", "addr", cfg.Address, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
}
