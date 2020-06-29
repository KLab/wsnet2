package auth

import (
	"crypto/sha256"
	"strconv"
	"time"
)

type Auth struct {
	psk string
}

func NewAuth(psk string) *Auth {
	return &Auth{psk}
}

func (a *Auth) ValidateHash(userId, nonce, hash string) bool {
	now := time.Now().Unix()
	for offset := 0; offset < 30; offset++ {
		h := a.GenerateHash(userId, now-int64(offset), nonce)
		if h == hash {
			return true
		}
	}
	return false
}

func (a *Auth) GenerateHash(userId string, timestamp int64, nonce string) string {
	s := sha256.New()
	s.Write([]byte(userId))
	s.Write([]byte(strconv.FormatInt(timestamp, 10)))
	s.Write([]byte(a.psk))
	s.Write([]byte(nonce))
	return string(s.Sum(nil))
}
