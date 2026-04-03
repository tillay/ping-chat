package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"encoding/json"
	"time"
)

type MsgRecord struct {
	LastMsgIp        string
	LastMsgEncrypted []byte
	LastMsgTimestamp int
}

var store = map[string]MsgRecord{}

// reply to any incoming pings
// gets incoming message ip and encrypted bytes from the client, returns what to reply with
func sendReply(ip string, incomingPayload []byte) string {
	key := extractHash(incomingPayload)
	if len(incomingPayload) == 32 {
		record, exists := store[string(key)]
		if !exists {
			record = MsgRecord{
				LastMsgIp:        ip,
				LastMsgEncrypted: []byte{},
				LastMsgTimestamp: int(time.Now().Unix()),
			}
			store[string(key)] = record
		}

		recordJson, err := json.Marshal(record)
		fmt.Println(string(recordJson))
		if err != nil {
			fmt.Println("Error encoding record:", err)
			return ""
		}
		return string(encryptUsingHash(recordJson, key))
	}

	updateChat(ip, incomingPayload)
	record := store[string(key)]
	recordJson, err := json.Marshal(record)
	if err != nil {
		fmt.Println("Error encoding record:", err)
		return ""
	}
	return string(encryptUsingHash(recordJson, key))
}

func runServer() {
	//flag.Parse()
	if *reauthKernel {
		enableKernelReplies(true)
	} else if *server {
		enableKernelReplies(false)
		listenForPackets()
	} else {
		fmt.Println(sendString(os.Args[1], "91.98.131.91"))
	}
}

func updateChat(ip string, encrypted []byte) {
	key := hex.EncodeToString(encrypted[:32])
	store[key] = MsgRecord{
		LastMsgIp:        ip,
		LastMsgEncrypted: encrypted,
		LastMsgTimestamp: int(time.Now().Unix()),
	}
}
