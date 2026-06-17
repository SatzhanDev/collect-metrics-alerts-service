// Package hashutil предоставляет утилиты для вычисления контрольных сумм.
package hashutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHash вычисляет HMAC-SHA256 от body с ключом key.
// Возвращает пустую строку, если key не задан.
func ComputeHash(body []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
