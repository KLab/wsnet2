package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"

	"golang.org/x/xerrors"
)

const (
	AllowedTimeGain = time.Second * 10
)

func ValidHMAC(mac, key []byte, args ...[]byte) bool {
	mac2 := CalculateHMAC(key, args...)
	return hmac.Equal(mac, mac2)
}

func CalculateHMAC(key []byte, args ...[]byte) []byte {
	mac := hmac.New(sha256.New, key)
	for _, arg := range args {
		mac.Write(arg)
	}
	return mac.Sum(nil)
}

// ValidAuthData validates authData.
// authData: base64 encoded [64bit nonce, 64bit timestamp, 256bit hmac]
func ValidAuthData(authData, key, userId string, expired time.Time) error {
	data, err := ValidAuthDataHash(authData, key, userId)
	if err != nil {
		return err
	}

	timedata := data[8:16]
	unixtime := binary.BigEndian.Uint64(timedata)
	timestamp := time.Unix(int64(unixtime), 0)

	if timestamp.After(time.Now().Add(AllowedTimeGain)) {
		return xerrors.Errorf("future timestamp: %v", timestamp)
	}

	if timestamp.Before(expired) {
		return xerrors.Errorf("expired: %v", timestamp)
	}

	return nil
}

// ValidAuthdataHash validate authdata hash.
// This function does not check the timestamp in authdata.
func ValidAuthDataHash(authData, key, userId string) ([]byte, error) {
	d, err := base64.StdEncoding.DecodeString(authData)
	if err != nil {
		return nil, xerrors.Errorf("decode base64: %w", err)
	}
	if len(d) != 8+8+32 {
		return nil, xerrors.Errorf("too short: %v", len(d))
	}

	data, hmac := d[:16], d[16:]

	if !ValidHMAC(hmac, []byte(key), []byte(userId), data) {
		return nil, xerrors.Errorf("hmac mismatch")
	}

	return data, nil
}

// GenerateAuthData generates base64 encoded authdata.
func GenerateAuthData(key, userId string, now time.Time) (string, error) {
	d := make([]byte, 8+8+32)

	nonce := d[0:8]
	n, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}
	if n != len(nonce) {
		return "", xerrors.Errorf("nonce length: %v", n)
	}

	// timestamp
	timestamp := d[8:16]
	binary.BigEndian.PutUint64(timestamp, uint64(now.Unix()))

	// hmac
	hmac := CalculateHMAC([]byte(key), []byte(userId), nonce, timestamp)
	if len(hmac) != 32 {
		return "", xerrors.Errorf("hmac length: %v", len(hmac))
	}
	copy(d[16:], hmac)

	return base64.StdEncoding.EncodeToString(d), nil
}
