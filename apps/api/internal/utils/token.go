package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func NewID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("utils: generate id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func NewToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("utils: generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HMACSHA256Hex(raw, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}
