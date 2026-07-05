package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

// HashString creates a deterministic hash value of a string
func HashString(val string) string {
	normalized := strings.TrimSpace(strings.ToLower(val))

	hashKey := []byte(os.Getenv("HMAC_KEY"))
	h := hmac.New(sha256.New, hashKey)
	h.Write([]byte(normalized))

	return hex.EncodeToString(h.Sum(nil))
}
