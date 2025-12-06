package crypto_test

import (
	"testing"

	"github.com/vibeswithkk/paylink/internal/crypto"
)

func TestComputeHMAC(t *testing.T) {
	key := "secret-key"
	data := "hello world"

	result := crypto.ComputeHMAC(key, data)

	if result == "" {
		t.Error("Expected non-empty HMAC result")
	}

	if len(result) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Expected 64 char hex string, got %d chars", len(result))
	}

	// Verify consistency
	result2 := crypto.ComputeHMAC(key, data)
	if result != result2 {
		t.Error("HMAC should be deterministic")
	}
}

func TestComputeHMACDifferentInputs(t *testing.T) {
	key := "secret-key"

	result1 := crypto.ComputeHMAC(key, "data1")
	result2 := crypto.ComputeHMAC(key, "data2")

	if result1 == result2 {
		t.Error("Different data should produce different HMACs")
	}
}

func BenchmarkComputeHMAC(b *testing.B) {
	key := "benchmark-key"
	data := "benchmark-data-payload"

	for i := 0; i < b.N; i++ {
		crypto.ComputeHMAC(key, data)
	}
}
