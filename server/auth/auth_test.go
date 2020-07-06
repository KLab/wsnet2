package auth

import (
	"strconv"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	userId := "alice"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	psk := "hoge"
	nonce := "1234"

	hash := GenerateHash(userId, timestamp, psk, nonce)
	if err := ValidateHash(userId, timestamp, psk, nonce, hash); err != nil {
		t.Fatalf("invalid hash: ")
	}

	timestamp = strconv.FormatInt(time.Now().Unix()-5, 10)
	hash = GenerateHash(userId, timestamp, psk, nonce)
	if err := ValidateHash(userId, timestamp, psk, nonce, hash); err != nil {
		t.Fatalf("invalid hash: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}

	// タイムスタンプ有効範囲外
	timestamp = strconv.FormatInt(time.Now().Unix()-(expirationTime+1), 10)
	hash = GenerateHash(userId, timestamp, psk, nonce)
	if err := ValidateHash(userId, timestamp, psk, nonce, hash); err == nil {
		t.Fatalf("Unexpected results: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}

	// 未来のタイムスタンプ
	timestamp = strconv.FormatInt(time.Now().Unix()+2, 10)
	hash = GenerateHash(userId, timestamp, psk, nonce)
	if err := ValidateHash(userId, timestamp, psk, nonce, hash); err == nil {
		t.Fatalf("Unexpected results: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}
}
