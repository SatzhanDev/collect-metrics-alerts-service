package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config/db"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/handler"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/middleware"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/file"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/postgres"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	cfg := parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()

	var (
		storage     repository.Storage
		fileStorage repository.FileStorage
		dbConn      *sql.DB
	)

	if cfg.DBDSN != "" {
		dbCfg := db.Config{
			DSN:             cfg.DBDSN,
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    defaultMaxIdleConns,
			ConnMaxLifetime: defaultConnMaxLifetime,
		}

		var err error
		dbConn, err = db.NewPostgres(dbCfg)
		if err != nil {
			logger.Log.Error("failed to connect to database", zap.Error(err))
			log.Fatal(err)
		}

		if err := db.RunMigrations(dbConn, "./migrations"); err != nil {
			logger.Log.Error("failed to run migrations", zap.Error(err))
			log.Fatal(err)
		}

		storage = postgres.New(dbConn)

	} else if cfg.FileStoragePath != "" {

		memStorage := mem.NewMemStorage()
		storage = memStorage

		fileStorage = file.NewJSONFileStorage()

	} else {
		storage = mem.NewMemStorage()
	}

	svc := service.NewMetricsService(storage, fileStorage, cfg)
	if cfg.DBDSN == "" && cfg.FileStoragePath != "" && cfg.Restore {
		if err := svc.RestoreFromFile(context.Background()); err != nil {
			logger.Log.Error("failed to restore metrics from file", zap.Error(err))
			log.Fatal(err)
		}
	}
	if cfg.DBDSN == "" && fileStorage != nil && cfg.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(cfg.StoreInterval)
			defer ticker.Stop()

			for range ticker.C {
				if err := svc.SaveToFile(context.Background()); err != nil {
					logger.Log.Error("failed to save metrics to file", zap.Error(err))
				}
			}
		}()
	}

	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.Use(logger.WithLogging)

	r.Get("/", h.GetList)
	r.Get("/value/{type}/{name}", h.Value)
	r.Post("/update/{type}/{name}/{value}", h.Update)

	r.Post("/update", h.UpdateJSON)
	r.Post("/update/", h.UpdateJSON)

	r.Post("/value", h.ValueJSON)
	r.Post("/value/", h.ValueJSON)

	r.Post("/updates", h.UpdateBatch)
	r.Post("/updates/", h.UpdateBatch)

	r.Get("/ping", h.Ping)

	addr := normalizeAddr(cfg.Addr)

	logger.Log.Info("Running server", zap.String("address", cfg.Addr))
	if err := http.ListenAndServe(addr, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

}
func normalizeAddr(in string) string {
	host, port, err := net.SplitHostPort(in)
	if err != nil {
		return in
	}
	if host == "localhost" {
		return ":" + port
	}
	return in
}
