package main

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

func deriveKey(pass []byte) []byte {
	hash := sha256.Sum256(pass)
	key, _ := scrypt.Key(pass, hash[:8], 32768, 8, 1, 32)
	return key
}

func encryptToBytes(data []byte, pass []byte) []byte {
	key := deriveKey(pass)
	aead, _ := chacha20poly1305.New(key)
	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)
	ct := aead.Seal(nonce, nonce, data, nil)
	return ct
}

func extractHash(data []byte) []byte {
	return data[:32]
}

func passHash(pass string) []byte {
	hash := sha256.Sum256([]byte(pass))
	return hash[:]
}

// this works fine
func decryptFromBytes(ct []byte, pass []byte) []byte {
	key := deriveKey(pass)
	if len(ct) == 0 {
		return []byte{}
	}
	aead, _ := chacha20poly1305.New(key)
	ns := aead.NonceSize()
	nonce, data := ct[:ns], ct[ns:]
	bytes, _ := aead.Open(nil, nonce, data, nil)
	return bytes
}

func decryptUsingPass(data []byte, pass string) string {
	if len(data) < 32 {
		return ""
	}
	key := deriveKey([]byte(pass))
	aead, _ := chacha20poly1305.New(key)
	ns := aead.NonceSize()
	if len(data) < ns {
		return "too short"
	}
	plain, err := aead.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return err.Error()
	}
	return string(plain)
}
