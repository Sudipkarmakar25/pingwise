package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// AES-256-GCM with a key loaded from config. The key version is stored next to every
// ciphertext, so a new key can be introduced later without breaking existing values.
var (
	currentKey     []byte
	currentVersion int
)

func Init(base64Key string, version int) error {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return fmt.Errorf("encryption key is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	if version <= 0 {
		return errors.New("encryption key version must be greater than 0")
	}
	currentKey = key
	currentVersion = version
	return nil
}

// Encrypt returns nonce+ciphertext and the version of the key used
func Encrypt(plaintext string) ([]byte, int, error) {
	gcm, err := newGCM(currentKey)
	if err != nil {
		return nil, 0, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, 0, err
	}
	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), currentVersion, nil
}

func Decrypt(ciphertext []byte, version int) (string, error) {
	if version != currentVersion {
		return "", fmt.Errorf("no encryption key for version %d", version)
	}
	gcm, err := newGCM(currentKey)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, sealed := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if key == nil {
		return nil, errors.New("encryption is not initialised")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
