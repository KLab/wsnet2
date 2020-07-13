package auth

import (
	"strconv"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	userId := "alice"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	key := "hoge"
	nonce, _ := GenerateNonce()

	mac := CalculateHMAC([]byte(key), userId, timestamp, nonce)
	if !ValidHMAC(mac, []byte(key), userId, timestamp, nonce) {
		t.Fatalf("invalid hmac")
	}
	hm := CalculateHexHMAC([]byte(key), userId, timestamp, nonce)
	if !ValidHexHMAC(hm, []byte(key), userId, timestamp, nonce) {
		t.Fatalf("invalid hex hmac")
	}

	invalidKey := "fuga"
	if ValidHMAC(mac, []byte(invalidKey), userId, timestamp, nonce) {
		t.Fatalf("invalid result")
	}
	if ValidHexHMAC(hm, []byte(invalidKey), userId, timestamp, nonce) {
		t.Fatalf("invalid result")
	}

	invalidUserId := "bob"
	if ValidHMAC(mac, []byte(key), invalidUserId, timestamp, nonce) {
		t.Fatalf("invalid result")
	}
	if ValidHexHMAC(hm, []byte(key), invalidUserId, timestamp, nonce) {
		t.Fatalf("invalid result")
	}

	invalidTimestamp := strconv.FormatInt(time.Now().Unix()+30, 10)
	if ValidHMAC(mac, []byte(key), userId, invalidTimestamp, nonce) {
		t.Fatalf("invalid result")
	}
	if ValidHexHMAC(hm, []byte(key), userId, invalidTimestamp, nonce) {
		t.Fatalf("invalid result")
	}

	invalidNonce, _ := GenerateNonce()
	if ValidHMAC(mac, []byte(key), userId, timestamp, invalidNonce) {
		t.Fatalf("invalid result")
	}
	if ValidHexHMAC(hm, []byte(key), userId, timestamp, invalidNonce) {
		t.Fatalf("invalid result")
	}
}
