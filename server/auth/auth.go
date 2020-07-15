package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func ValidHMAC(mac, key []byte, args ...string) bool {
	mac2 := CalculateHMAC(key, args...)
	return hmac.Equal(mac, mac2)
}

func ValidHexHMAC(hm string, key []byte, args ...string) bool {
	mac, err := hex.DecodeString(hm)
	if err != nil {
		return false
	}
	return ValidHMAC(mac, key, args...)
}

func CalculateHMAC(key []byte, args ...string) []byte {
	mac := hmac.New(sha256.New, []byte(key))
	for _, arg := range args {
		mac.Write([]byte(arg))
	}
	return mac.Sum(nil)
}

func CalculateHexHMAC(key []byte, args ...string) string {
	return hex.EncodeToString(CalculateHMAC(key, args...))
}

func GenerateNonce() (string, error) {
	buf := make([]byte, 8)
	n, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:n]), nil
}
