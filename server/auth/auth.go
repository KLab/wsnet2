package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"golang.org/x/xerrors"
)

const expirationTime = 30

func ValidateHash(userId, timestamp, psk, nonce, hash string) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return xerrors.Errorf("invalid timestamp: %w", err)
	}
	now := time.Now().Unix()
	if now < ts {
		return xerrors.Errorf("invalid timestamp: now=%v, ts=%v", now, ts)
	}
	if now-ts > expirationTime {
		return xerrors.Errorf("expired timestamp: now=%v, ts=%v", now, ts)
	}
	h := GenerateHash(userId, timestamp, psk, nonce)
	if h != hash {
		return xerrors.New("invalid hash")
	}
	return nil
}

func GenerateHash(userId, timestamp, psk, nonce string) string {
	s := sha256.New()
	s.Write([]byte(userId))
	s.Write([]byte(timestamp))
	s.Write([]byte(psk))
	s.Write([]byte(nonce))
	return hex.EncodeToString(s.Sum(nil))
}

func GenerateNonce() (string, error) {
	buf := make([]byte, 8)
	n, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:n]), nil
}
