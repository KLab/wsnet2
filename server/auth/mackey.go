package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"strings"

	"golang.org/x/xerrors"
)

// DecryptMACKey decodes a MACKey
func DecryptMACKey(encMKey, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encMKey)
	if err != nil {
		return "", err
	}

	ckey := sha256.Sum256([]byte(key))
	b, err := aes.NewCipher(ckey[:])
	if err != nil {
		return "", err
	}

	bs := b.BlockSize()
	iv := data[:bs]
	dst := make([]byte, len(data)-bs)

	cipher.NewCBCDecrypter(b, iv).CryptBlocks(dst, data[bs:])

	return strings.TrimRight(string(dst), "\x00"), nil
}

// EncryptMAckey encrypts macKey and returns base64 string
func EncryptMACKey(key, macKey string) (string, error) {
	ckey := sha256.Sum256([]byte(key))
	b, err := aes.NewCipher(ckey[:])
	if err != nil {
		return "", err
	}

	bs := b.BlockSize()
	blocks := int((len(macKey) + bs - 1) / bs)
	src := make([]byte, blocks*bs)
	if n := copy(src, []byte(macKey)); n != len(macKey) {
		return "", xerrors.Errorf("copy macKey to block: %v, %v", n, len(macKey))
	}

	buf := make([]byte, bs+blocks*bs) // iv + dst

	// iv
	iv := buf[:bs]
	if n, err := rand.Read(iv); err != nil {
		return "", err
	} else if n != bs {
		return "", xerrors.Errorf("IV length %v", n)
	}

	cipher.NewCBCEncrypter(b, iv).CryptBlocks(buf[bs:], src)

	return base64.StdEncoding.EncodeToString(buf), nil
}

func ValidateMsgHMAC(mac hash.Hash, data []byte) ([]byte, bool) {
	hs := mac.Size()
	if len(data) < hs {
		return nil, false
	}

	h := data[len(data)-hs:]
	data = data[:len(data)-hs]

	mac.Write(data)
	sum := mac.Sum(nil)
	mac.Reset()

	return data, hmac.Equal(h, sum)
}
