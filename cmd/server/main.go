package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/audit"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/buildinfo"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config/db"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/cryptoutil"
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
	fmt.Printf("Build version: %s\n", buildinfo.NA(buildinfo.Version))
	fmt.Printf("Build date: %s\n", buildinfo.NA(buildinfo.Date))
	fmt.Printf("Build commit: %s\n", buildinfo.NA(buildinfo.Commit))

	cfg := parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()

	go func() {
		logger.Log.Info("pprof server running", zap.String("address", "localhost:6060"))

		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			logger.Log.Error("pprof server error", zap.Error(err))
		}
	}()

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
				saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := svc.SaveToFile(saveCtx); err != nil {
					logger.Log.Error("failed to save metrics to file", zap.Error(err))
				}
				cancel()
			}
		}()
	}

	var privKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		var err error
		privKey, err = cryptoutil.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			logger.Log.Error("failed to load private key", zap.Error(err))
			log.Fatal(err)
		}
	}

	auditPublisher := audit.NewPublisher(logger.Log)
	if cfg.AuditFile != "" {
		auditPublisher.Subscribe(audit.NewFileObserver(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditPublisher.Subscribe(audit.NewHTTPObserver(cfg.AuditURL))
	}

	h := handler.NewMetricsHandler(svc, auditPublisher)

	trustedSubnetMiddleware, err := middleware.TrustedSubnetMiddleware(cfg.TrustedSubnet)
	if err != nil {
		logger.Log.Error("invalid trusted subnet", zap.Error(err))
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(trustedSubnetMiddleware)
	r.Use(middleware.CryptoMiddleware(privKey))
	r.Use(middleware.HashValidationMiddleware(cfg.Key))
	r.Use(middleware.GzipMiddleware)
	r.Use(logger.WithLogging)
	r.Use(middleware.HashResponseMiddleware(cfg.Key))

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
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		logger.Log.Info("Running server", zap.String("address", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("server shutdown error", zap.Error(err))
	}

	if fileStorage != nil {
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := svc.SaveToFile(saveCtx)
		saveCancel()
		if err != nil {
			logger.Log.Error("failed to save metrics on shutdown", zap.Error(err))
		}
	}

	if dbConn != nil {
		if err := dbConn.Close(); err != nil {
			logger.Log.Error("failed to close db connection", zap.Error(err))
		}
	}

	logger.Log.Info("Server stopped")
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
