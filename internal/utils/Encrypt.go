package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

// getAESKey loads and decodes the AES key from env
func getAESKey() ([]byte, error) {
	keyB64 := GetEnvVariables().AESKey
	rawKey, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, err
	}
	if l := len(rawKey); l != 16 && l != 24 && l != 32 {
		return nil, errors.New("invalid AES key size, must be 16, 24, or 32 bytes after base64 decoding")
	}
	return rawKey, nil
}

// encrypt with AES-GCM and return base64 url-safe string
func EncodeWithSecret(plaintext string) (string, error) {
	rawKey, err := getAESKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(rawKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt back to plaintext
func DecodeWithSecret(encoded string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	rawKey, err := getAESKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(rawKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// concat multiple values with ":" delimiter
func ConcatQueryParams(args ...string) string {
	return strings.Join(args, ":")
}

// split back into parts
func SplitQueryParams(str string) []string {
	return strings.Split(str, ":")
}
