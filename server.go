package main

import (
	"encoding/json"
	"fmt"
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
func sendReply(ip string, incomingPayload []byte) []byte {
	key := extractHash(incomingPayload)
	if len(incomingPayload) == 32 {
		record, exists := store[string(key)]
		if !exists {
			record = MsgRecord{
				LastMsgIp:        ip,
				LastMsgEncrypted: key,
				LastMsgTimestamp: int(time.Now().Unix()),
			}
			store[string(key)] = record
		}

		recordJson, err := json.Marshal(record)
		if err != nil {
			fmt.Println("Error encoding record:", err)
			return []byte{}
		}
		return encryptUsingHash(recordJson, key)
	}
	fmt.Println("OMG INCOMING MESSAGE")

	record, exists := store[string(key)]
	if !exists {
		fmt.Println("AYO I GOT HERE NOT GOOD")
		record = MsgRecord{
			LastMsgIp:        ip,
			LastMsgEncrypted: key,
			LastMsgTimestamp: int(time.Now().Unix()),
		}
		store[string(key)] = record
	}
	fmt.Println(record.LastMsgEncrypted)
	recordJson, err := json.Marshal(record)
	fmt.Println("This is running. Told you so.")
	fmt.Println(string(recordJson))
	if err != nil {
		fmt.Println("Error encoding record:", err)
		return []byte{}
	}
	return encryptUsingHash(recordJson, key)
}

func runServer() {
	listenForPackets()
}
