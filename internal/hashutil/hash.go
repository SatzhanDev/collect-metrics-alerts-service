package hashutil

import (
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHash(body []byte, key string) string {
	if key == "" {
		return ""
	}
	sum := sha256.Sum256(append(body, []byte(key)...))
	return hex.EncodeToString(sum[:])
}
