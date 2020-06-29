package auth

import (
	"testing"
	"time"
)

const psk = "hoge"

func TestAuth(t *testing.T) {
	userId := "alice"
	timestamp := time.Now().Unix()
	nonce := "1234"

	auth := NewAuth(psk)
	hash := auth.GenerateHash(userId, timestamp, nonce)
	if !auth.ValidateHash(userId, nonce, hash) {
		t.Fatalf("invalid hash: ")
	}

	timestamp = time.Now().Unix() - 5
	hash = auth.GenerateHash(userId, timestamp, nonce)
	if !auth.ValidateHash(userId, nonce, hash) {
		t.Fatalf("invalid hash: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}

	// タイムスタンプ有効範囲外
	timestamp = time.Now().Unix() - 31
	hash = auth.GenerateHash(userId, timestamp, nonce)
	if auth.ValidateHash(userId, nonce, hash) {
		t.Fatalf("Unexpected results: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}

	// 未来のタイムスタンプ
	timestamp = time.Now().Unix() + 2
	hash = auth.GenerateHash(userId, timestamp, nonce)
	if auth.ValidateHash(userId, nonce, hash) {
		t.Fatalf("Unexpected results: userId=%v, timestamp=%v, nonce=%v, hash=%v", userId, timestamp, nonce, hash)
	}
}
