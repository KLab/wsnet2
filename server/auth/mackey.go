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
func DecryptMACKey(appKey, encMKey string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encMKey)
	if err != nil {
		return "", err
	}

	ckey := sha256.Sum256([]byte(appKey))
	b, err := aes.NewCipher(ckey[:])
	if err != nil {
		return "", err
	}

	bs := b.BlockSize()
	if len(data) < bs*2 {
		return "", xerrors.Errorf("data too short")
	}
	iv := data[:bs]
	dst := make([]byte, len(data)-bs)

	cipher.NewCBCDecrypter(b, iv).CryptBlocks(dst, data[bs:])

	return strings.TrimRight(string(dst), "\x00"), nil
}

// EncryptMAckey encrypts macKey and returns base64 string
func EncryptMACKey(appKey, macKey string) (string, error) {
	ckey := sha256.Sum256([]byte(appKey))
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

// ValidateMsgHMAC validates the hmac of a websocket message.
func ValidateMsgHMAC(mac hash.Hash, data []byte) ([]byte, error) {
	dlen := len(data) - mac.Size()
	if dlen < 0 {
		return nil, xerrors.Errorf("data=[% X]", data)
	}
	data, h := data[:dlen], data[dlen:]
	hash := CalculateMsgHMAC(mac, data)
	if !hmac.Equal(h, hash) {
		return nil, xerrors.Errorf("hash=(%v, %v)", h, hash)
	}
	return data, nil
}

func CalculateMsgHMAC(mac hash.Hash, data []byte) []byte {
	mac.Write(data)
	defer mac.Reset()
	return mac.Sum(nil)
}

func GenMACKey() string {
	buf := make([]byte, 16/4*3)
	rand.Read(buf)
	return base64.StdEncoding.EncodeToString(buf)
}
