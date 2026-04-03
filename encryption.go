package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"//if throwing errors, use "go get golang.com/..."
	"golang.org/x/crypto/scrypt"
)

func runClient() {//takes message and password, encrypts, decrypts, checks hash
	var pass string//please note servers will use main at main.go
	var message string

	
	fmt.Println("Input password")
	fmt.Scan(&pass)
	fmt.Println("Input message")
	fmt.Scan(&message)

	encrypted := encrypt(message, pass)
	fmt.Println(encrypted)
	fmt.Println(decrypt(encrypted, pass))
	fmt.Println(extractHash(encrypted))
}

func deriveKey(pass string) []byte { //used in encrypt and decrypt
	hash := sha256.Sum256([]byte(pass))
	key, _ := scrypt.Key([]byte(pass), hash[:8], 32768, 8, 1, 32)
	return key
}

func encrypt(text, pass string) []byte {//generate key and hash
	hash := sha256.Sum256([]byte(pass))
	key := deriveKey(pass)
	aead, _ := chacha20poly1305.New(key)
	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)
	ct := aead.Seal(nonce, nonce, []byte(text), nil)
	return append(hash[:], ct...) //encrypted in form of slice of bytes 
}

func extractHash(data []byte) []byte {
	return data[:32]
}

func decrypt(data []byte, pass string) string {
	key := deriveKey(pass)
	aead, _ := chacha20poly1305.New(key)
	ct := data[32:]
	ns := aead.NonceSize()
	plain, err := aead.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return ""
	}
	return string(plain)
}