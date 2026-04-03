package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type MsgRecord struct {
	LastMsgIp        string
	LastMsgEncrypted []byte
	LastMsgTimestamp int
	LastMsgHash      []byte
}

var store = map[string]MsgRecord{}

// reply to any incoming pings
// gets incoming message ip and text, returns what to reply with
func sendReply(ip string, text string) string {
	return ip + ", i got your " + text + ". thx"
}

func updateChat(ip string, encrypted []byte) {
	key := hex.EncodeToString(encrypted[:32])
	store[key] = MsgRecord{
		LastMsgIp:        ip,
		LastMsgEncrypted: encrypted,
		LastMsgTimestamp: int(time.Now().Unix()),
		LastMsgHash:      func() []byte { h := sha256.Sum256(encrypted); return h[:] }(),
	}
}

func runServer() {
	//


}
