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

func encryptUsingPass(text string, pass string) []byte {
	//fmt.Println("encrypting with pass " + string(pass) + " and text " + text)
	hash := sha256.Sum256([]byte(pass))
	key := deriveKey([]byte(pass))
	aead, _ := chacha20poly1305.New(key)
	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)
	ct := aead.Seal(nonce, nonce, []byte(text), nil)
	//fmt.Println("it became" + string(append(hash[:], ct...)))
	return append(hash[:], ct...)
}

func encryptUsingHash(data []byte, pass []byte) []byte {
	//fmt.Println("encrypting with hash " + string(pass) + " and data " + string(data))
	key := deriveKey(pass)
	aead, _ := chacha20poly1305.New(key)
	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)
	ct := aead.Seal(nonce, nonce, data, nil)
	//fmt.Println("it became" + string(ct))
	return ct
}

func extractHash(data []byte) []byte {
	return data[:32]
}

func passHash(pass string) []byte {
	hash := sha256.Sum256([]byte(pass))
	return hash[:]
}

func decryptUsingPass(data []byte, pass string) string {
	//fmt.Println("decrypting with pass " + string(pass) + " and data " + string(data))
	if len(data) < 32 {
		return ""
	}
	key := deriveKey([]byte(pass))
	aead, _ := chacha20poly1305.New(key)
	ct := data[32:]
	ns := aead.NonceSize()
	if len(ct) < ns {
		return "too short"
	}
	plain, err := aead.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return err.Error()
	}
	return string(plain)
}

func decryptUsingHash(ct []byte, pass []byte) []byte {
	//fmt.Println("decrypting with hash " + string(pass) + " and data " + string(ct))
	key := deriveKey(pass)
	aead, _ := chacha20poly1305.New(key)
	ns := aead.NonceSize()
	nonce, data := ct[:ns], ct[ns:]
	bytes, _ := aead.Open(nil, nonce, data, nil)
	//fmt.Println("it became" + string(bytes))
	return bytes
}
