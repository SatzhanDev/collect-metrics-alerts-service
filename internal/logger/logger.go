// Package logger предоставляет глобальный логгер и HTTP-middleware для логирования запросов.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: http.StatusOK,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		if responseData.status >= 500 {
			Log.Error("request failed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", responseData.status),
				zap.Duration("duration", duration),
				zap.Int("size", responseData.size),
			)
		} else {
			Log.Info("request completed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", responseData.status),
				zap.Duration("duration", duration),
				zap.Int("size", responseData.size),
			)
		}
	}
	return http.HandlerFunc(logFn)
}
