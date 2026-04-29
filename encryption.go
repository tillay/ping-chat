package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"

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

// returns convoHash, senderHash
func extractHashes(data []byte) ([]byte, []byte) {
	return deobfuscate(data[:16]), deobfuscate(data[16:32])
}

func passHash(pass string) []byte {
	hash := sha256.Sum256([]byte(pass))
	return hash[:16]
}

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
	if len(data) < 16 {
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

func decryptUserBlob(blob []byte) *userInfo {
	plain := decryptUsingPass(blob, *pass)
	var ub UserBlob
	if err := json.Unmarshal([]byte(plain), &ub); err != nil {
		return nil
	}
	return &userInfo{User: ub.User, Color: ub.Color}
}

func obfuscate(hash []byte) []byte {
	nonce := make([]byte, 1)
	rand.Read(nonce)
	out := make([]byte, len(hash))
	out[0] = nonce[0]
	x := uint32(nonce[0]) ^ 0x5A827999
	for i := 1; i < len(hash); i++ {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		out[i] = hash[i] ^ byte(x)
	}
	return out
}

func deobfuscate(packet []byte) []byte {
	nonce := packet[0]
	out := make([]byte, len(packet))
	x := uint32(nonce) ^ 0x5A827999
	for i := 1; i < len(packet); i++ {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		out[i] = packet[i] ^ byte(x)
	}
	return out
}
