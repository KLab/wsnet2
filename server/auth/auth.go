package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/xerrors"
)

func ValidHMAC(mac, key []byte, args ...[]byte) bool {
	mac2 := CalculateHMAC(key, args...)
	return hmac.Equal(mac, mac2)
}

func ValidHexHMAC(hm string, key []byte, args ...[]byte) bool {
	mac, err := hex.DecodeString(hm)
	if err != nil {
		return false
	}
	return ValidHMAC(mac, key, args...)
}

func CalculateHMAC(key []byte, args ...[]byte) []byte {
	mac := hmac.New(sha256.New, key)
	for _, arg := range args {
		mac.Write(arg)
	}
	return mac.Sum(nil)
}

func CalculateHexHMAC(key []byte, args ...[]byte) string {
	return hex.EncodeToString(CalculateHMAC(key, args...))
}

func GenerateNonce() ([]byte, error) {
	buf := make([]byte, 8)
	n, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// ValidAuthData validates authData.
// authData: base64 encoded [64bit nonce, 64bit timestamp, 256bit hmac]
func ValidAuthData(authData, key, userId string, expired time.Time) error {
	d, err := base64.StdEncoding.DecodeString(authData)
	if err != nil {
		return xerrors.Errorf("invalid authdata: %w", err)
	}
	if len(d) != 8+8+32 {
		return xerrors.Errorf("invalid authdata: length=%v", len(d))
	}

	nonce, timedata, hmac := d[:8], d[8:16], d[16:]

	if !ValidHMAC(hmac, []byte(key), []byte(userId), nonce, timedata) {
		return xerrors.Errorf("invalid authdata: hmac mismatch")
	}

	unixtime := binary.BigEndian.Uint64(timedata)
	timestamp := time.Unix(int64(unixtime), 0)
	fmt.Printf("--- unix: %v\ntime: %v\nlimit:%v\n", unixtime, timestamp, expired)

	if timestamp.Before(expired) {
		return xerrors.Errorf("invalid authdata: expired")
	}

	return nil
}

// GenerateAuthData generates base64 encoded authdata.
func GenerateAuthData(key, userId string, now time.Time) (string, error) {
	d := make([]byte, 8+8+32)

	nonce := d[0:8]
	n, err := GenerateNonce()
	if err != nil {
		return "", err
	}
	copy(nonce, n)

	// timestamp
	timestamp := d[8:16]
	binary.BigEndian.PutUint64(timestamp, uint64(now.Unix()))

	// hmac
	hmac := CalculateHMAC([]byte(key), []byte(userId), nonce, timestamp)
	if len(hmac) != 32 {
		return "", xerrors.Errorf("invalid hmac length %v", len(hmac))
	}
	copy(d[16:], hmac)

	return base64.StdEncoding.EncodeToString(d), nil
}
