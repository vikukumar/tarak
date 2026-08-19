// Package security — AES-256-GCM encryption for Secret objects.
//
// Secrets stored in Tarak's state store are encrypted at rest using AES-256-GCM.
// The encryption key is derived from a cluster master key that is stored
// separately from the state store database.
//
// Encrypted format:
//
//	[1 byte version][12 byte nonce][ciphertext][16 byte GCM tag]
//
// Version 0x01 = AES-256-GCM with random nonce.
//
// The master key is 32 bytes (256 bits) of random data.
// Key rotation is done by re-encrypting all secrets with a new key.
//
// Security properties:
//   - Each encryption uses a fresh random nonce.
//   - GCM tag provides authenticated encryption (integrity + confidentiality).
//   - The plaintext is never written to disk.
//   - Secret values are never logged (enforced by using []byte, not string, for data).
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

const (
	encryptionVersion = byte(0x01)
	keySize           = 32 // AES-256
	nonceSize         = 12 // GCM standard nonce
)

// Encryptor provides AES-256-GCM encryption and decryption.
type Encryptor struct {
	key []byte
}

// NewEncryptor creates an Encryptor with the given 32-byte AES-256 key.
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("encryption key must be exactly %d bytes, got %d", keySize, len(key))
	}
	keyCopy := make([]byte, keySize)
	copy(keyCopy, key)
	return &Encryptor{key: keyCopy}, nil
}

// GenerateEncryptionKey generates a new 32-byte random encryption key.
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}
	return key, nil
}

// LoadOrGenerateKey loads an encryption key from a file, or generates and saves a new one.
// The key file is created with mode 0600.
func LoadOrGenerateKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) != keySize {
			return nil, fmt.Errorf("encryption key file %q has wrong size: got %d, expected %d", path, len(data), keySize)
		}
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read encryption key %q: %w", path, err)
	}

	// Generate and save a new key.
	key, err := GenerateEncryptionKey()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("write encryption key %q: %w", path, err)
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// The returned ciphertext includes the version byte, nonce, and GCM tag.
// The plaintext slice is not mutated.
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Output: version || nonce || ciphertext+tag
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 1+nonceSize+len(ciphertext))
	out[0] = encryptionVersion
	copy(out[1:], nonce)
	copy(out[1+nonceSize:], ciphertext)

	return out, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// Returns the original plaintext or an error if authentication fails.
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 1+nonceSize+16 { // 16 = GCM tag size
		return nil, fmt.Errorf("ciphertext too short")
	}

	version := ciphertext[0]
	if version != encryptionVersion {
		return nil, fmt.Errorf("unknown encryption version: 0x%02x", version)
	}

	nonce := ciphertext[1 : 1+nonceSize]
	encrypted := ciphertext[1+nonceSize:]

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: authentication failed")
	}
	return plaintext, nil
}

// IsEncrypted reports whether data looks like Tarak-encrypted ciphertext.
func IsEncrypted(data []byte) bool {
	return len(data) > 0 && data[0] == encryptionVersion
}
