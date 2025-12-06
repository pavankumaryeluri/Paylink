package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHMAC computes HMAC-SHA256 and returns hex-encoded result
// This is a pure Go implementation for maximum portability
// In production with high throughput, consider the CGO version
func ComputeHMAC(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies an HMAC signature using constant-time comparison
func VerifyHMAC(key, data, signature string) bool {
	expected := ComputeHMAC(key, data)
	return hmac.Equal([]byte(expected), []byte(signature))
}
