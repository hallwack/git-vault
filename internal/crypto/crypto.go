// Package crypto provides a simple interface for encrypting and decrypting 
// data using the XChaCha20-Poly1305 AEAD cipher. It is designed to be used in 
// the context of Git Vault, but it is agnostic to Git or any specific file 
// format. The package ensures that the key and nonce sizes are correct and 
// handles the encryption and decryption processes, returning errors when necessary.
package crypto

import (
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// KeySize is the master key length required by XChaCha20-Poly1305.
const KeySize = chacha20poly1305.KeySize

// Encrypt seals plaintext using the given key and nonce, returning ciphertext
// with an appended authentication tag. This package has no knowledge of Git or
// files - it only transform bytes.
func Encrypt(key, nonce, plaintext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: got %d, want %d", len(key), KeySize)
	}

	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), NonceSize)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// dst=nil so Seal allocates a fresh slice; additionalData=nil since Git Vault
	// has no extra authenticated metadata to bind.
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: got %d, want %d", len(key), KeySize)
	}

	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), NonceSize)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
	}

	return plaintext, nil
}
