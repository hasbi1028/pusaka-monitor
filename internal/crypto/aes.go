package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

var encryptionKey []byte

// Init menginisialisasi encryption key dari config
func Init(key string) error {
	if len(key) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY minimal 32 karakter, dapat: %d", len(key))
	}
	encryptionKey = []byte(key[:32])
	log.Println("[CRYPTO] AES-GCM encryption diaktifkan untuk password Pusaka")
	return nil
}

// Encrypt mengenkripsi plaintext dengan AES-256-GCM
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" || len(encryptionKey) == 0 {
		return plaintext, nil
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt mendekripsi ciphertext AES-256-GCM
func Decrypt(encoded string) (string, error) {
	if encoded == "" || len(encryptionKey) == 0 {
		return encoded, nil
	}
	// Kalau tidak ter-encrypt, return as-is
	if !IsEncrypted(encoded) {
		return encoded, nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded, nil // return as-is kalau decode gagal
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return encoded, nil
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return encoded, nil
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return encoded, nil
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encoded, nil
	}
	return string(plaintext), nil
}

// IsEncrypted mendeteksi apakah string sudah ter-encrypt
// AES-GCM ciphertext + nonce + auth tag minimum 12+16=28 bytes
// base64 dari 28 bytes = ~40 characters
func IsEncrypted(s string) bool {
	if len(s) < 40 {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	// Ciphertext harus >= 28 bytes (12 nonce + 16 tag) + minimal 1 byte data
	return len(decoded) >= 29
}
