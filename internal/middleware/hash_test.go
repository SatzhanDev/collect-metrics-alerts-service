package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/hashutil"
	"github.com/stretchr/testify/require"
)

func TestHashResponseMiddleware_AddsHash(t *testing.T) {
	h := HashResponseMiddleware("secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.NotEmpty(t, rec.Header().Get("HashSHA256"))
}
func TestHashResponseMiddleware_EmptyKey(t *testing.T) {
	h := HashResponseMiddleware("")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
func TestHashValidationMiddleware_ValidHash(t *testing.T) {
	body := "hello"

	hash := hashutil.ComputeHash([]byte(body), "secret")

	h := HashValidationMiddleware("secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(body),
	)

	req.Header.Set("HashSHA256", hash)

	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHashValidationMiddleware_InvalidHash(t *testing.T) {
	h := HashValidationMiddleware("secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("hello"),
	)

	req.Header.Set("HashSHA256", "wrong-hash")

	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHashValidationMiddleware_NoHash(t *testing.T) {
	h := HashValidationMiddleware("secret")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("hello"),
	)

	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
