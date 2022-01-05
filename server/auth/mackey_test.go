package auth

import "testing"

func TestMACKey(t *testing.T) {
	appkey := "testkey2"
	mackey := "testMACKey2"

	encMkey, err := EncryptMACKey(appkey, mackey)
	if err != nil {
		t.Fatalf("EncryptMACKey: %v", err)
	}

	r, err := DecryptMACKey(appkey, encMkey)
	if err != nil {
		t.Fatalf("DecryptMACKey: %v", err)
	}

	if r != mackey {
		t.Fatalf("decrypted = %q, wants %q", r, mackey)
	}
}
