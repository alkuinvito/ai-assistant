package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

// Sign returns a hex-encoded HMAC-SHA256 signature of message using secret.
// Unlike HashString, it does not normalize the input.
func Sign(secret, message string) (string, error) {
	if secret == "" {
		return "", errors.New("signing secret is not configured")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))

	return hex.EncodeToString(h.Sum(nil)), nil
}
