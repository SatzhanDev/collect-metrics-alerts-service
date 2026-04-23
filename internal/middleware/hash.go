package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/hashutil"
)

const maxHashBodySize = 1 << 20

type hashResponseWriter struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *hashResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func HashResponseMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			rw := &hashResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			next.ServeHTTP(rw, r)

			bodyBytes := rw.body.Bytes()
			hash := hashutil.ComputeHash(bodyBytes, key)

			if hash != "" {
				w.Header().Set("HashSHA256", hash)
			}

			w.WriteHeader(rw.status)
			_, _ = w.Write(bodyBytes)
		})
	}
}

func HashValidationMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			limited := io.LimitReader(r.Body, maxHashBodySize+1)
			body, err := io.ReadAll(limited)
			if err != nil {
				http.Error(w, "cannot read request body", http.StatusBadRequest)
				return
			}

			r.Body.Close()

			if len(body) > maxHashBodySize {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			expectedHash := hashutil.ComputeHash(body, key)
			receivedHash := r.Header.Get("HashSHA256")

			if receivedHash == "" {
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				next.ServeHTTP(w, r)
				return
			}
			if receivedHash != expectedHash {
				http.Error(w, "invalid hash", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(w, r)
		})
	}
}
