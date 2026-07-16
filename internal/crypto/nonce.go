// Package crypto provides cryptographic utilities for secure data handling, 
// including nonce generation for encryption algorithms like XChaCha20-Poly1305.
package crypto

import (
	"crypto/rand"
	"fmt"
)

// NonceSize is the nonce length required by XChaCha20-Poly1305. (24 bytes — 
// much larger than AES-GCM's 12 bytes, which is why a randomly generated nonce
// is safe to reuse-check against here: the odds of collision are astronomically 
// low even across many files and many commits.)
const NonceSize = 24

// GenerateNonce returns a new cryptographically random nonce. A fresh nonce 
// must be generated for every single encryption call — never reuse a nonce 
// with the same key.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return nonce, nil
}
