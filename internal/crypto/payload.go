// Package crypto provides cryptographic utilities for secure data handling,
// including nonce generation for encryption algorithms like XChaCha20-Poly1305.
package crypto

import "fmt"

// PayloadVersion identified the on-disk format of encrypted files. Bumping this
// lets future versions detect and reject (or migrate) payloads written by an
// older scheme
const PayloadVersion byte = 1

// minPayloadSize is the smallest a valid payload can be: 1 version byte + full
// nonce. Anything shorted cannot possibly be valid.
const minPayloadSize = 1 + NonceSize

// Serialize combines the version byte, nonce, and ciphertext into a single blob
// suitable for writing to disk as the encrypted file contents.
//
// Layout: [1 byte version][NonceSize bytes nonce][ciphertext...]
func Serialize(nonce, ciphertext []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), NonceSize)
	}

	payload := make([]byte, 0, minPayloadSize+len(ciphertext))
	payload = append(payload, PayloadVersion)
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)

	return payload, nil
}

// Deserialize splits a payload (as produced by Serialize) back into its nonce
// and ciphertext components, after validating the version byte and minimum length
func Deserialize(payload []byte) (nonce, ciphertext []byte, err error) {
	if len(payload) < minPayloadSize {
		return nil, nil, fmt.Errorf("payload too short: got %d bytes, want at least %d", len(payload), minPayloadSize)
	}

	version := payload[0]
	if version != PayloadVersion {
		return nil, nil, fmt.Errorf("unsupported payload version: got %d, want %d", version, PayloadVersion)
	}

	nonce = payload[1 : 1+NonceSize]
	ciphertext = payload[1+NonceSize:]

	return nonce, ciphertext, nil
}
