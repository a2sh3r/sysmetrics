// Package hash provides functions for calculating and verifying hashes.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateHash calculates a SHA256 hash for the given data and key.
func CalculateHash(data string, key string) string {
	if key == "" {
		return ""
	}

	h := sha256.New()

	h.Write([]byte(data + key))

	hashBytes := h.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

// VerifyHash verifies that the given hash matches the data and key.
func VerifyHash(data string, key string, hash string) error {
	if key == "" {
		return nil
	}

	calculatedHash := CalculateHash(data, key)
	if calculatedHash != hash {
		return fmt.Errorf("hash verification failed: expected %s, got %s", hash, calculatedHash)
	}

	return nil
}
