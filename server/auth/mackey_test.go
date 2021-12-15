package auth

import "testing"

func TestMACKey(t *testing.T) {
	key := "testkey2"
	mackey := "testMACKey2"

	encMkey, err := EncryptMACKey(mackey, key)
	if err != nil {
		t.Fatalf("EncryptMACKey: %v", err)
	}

	r, err := DecryptMACKey(encMkey, key)
	if err != nil {
		t.Fatalf("DecryptMACKey: %v", err)
	}

	if r != mackey {
		t.Fatalf("decrypted = %q, wants %q", r, mackey)
	}
}
