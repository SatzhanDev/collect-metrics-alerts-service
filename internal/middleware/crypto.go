package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/cryptoutil"
)

// CryptoMiddleware дешифрует тело запроса приватным ключом RSA priv перед
// передачей его дальше по цепочке обработчиков. Если priv не задан (nil),
// тело запроса не изменяется — предполагается, что агент не шифрует
// сообщения. Middleware должна выполняться до валидации хэша и распаковки
// gzip, так как агент вычисляет хэш и сжимает данные до их шифрования.
func CryptoMiddleware(priv *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if priv == nil {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			if len(body) == 0 {
				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
				return
			}

			decrypted, err := cryptoutil.Decrypt(priv, body)
			if err != nil {
				http.Error(w, "cannot decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			next.ServeHTTP(w, r)
		})
	}
}
