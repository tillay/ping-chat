package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type MsgRecord struct {
	MsgIp        string
	MsgPayload   []byte
	MsgTimestamp int
	IpLocation   string
}

var store = map[string]MsgRecord{}
var locationCache = map[string]string{}

// reply to any incoming pings
// gets incoming message ip and encrypted bytes from the client, returns what to reply with
func sendReply(ip string, incomingPayload []byte) []byte {
	key := extractHash(incomingPayload)
	if len(incomingPayload) == 32 {
		// this means that it's a routine ping for new messages
		// the 32 bytes is the hash of the password that the server should encrypt with
		record, exists := store[string(key)]
		if !exists {
			// create a new conversation struct entry for this password hash
			store[string(key)] = MsgRecord{MsgIp: ip, MsgTimestamp: int(time.Now().Unix())}
		} else {
			_ = record
		}
		record = store[string(key)]
		recordJson, _ := json.Marshal(record)
		return encryptToBytes(recordJson, key)
	}

	encrypted := incomingPayload[32:]

	// package together the ip of the sender, encrypted text sender sent, and timestamp
	store[string(key)] = MsgRecord{
		MsgIp:        ip,
		MsgPayload:   encrypted,
		MsgTimestamp: int(time.Now().Unix()),
		IpLocation:   getIpLocation(ip),
	}
	record := store[string(key)]
	recordJson, _ := json.Marshal(record)

	// encrypt and send package to client
	return encryptToBytes(recordJson, key)
}

func getIpLocation(ip string) string {
	if loc, ok := locationCache[ip]; ok {
		return loc
	}
	resp, err := http.Get("https://ipinfo.io/" + ip + "/json")
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()
	var data map[string]string
	json.NewDecoder(resp.Body).Decode(&data)
	loc := data["city"] + ", " + data["country"]
	locationCache[ip] = loc
	return loc
}

func runServer() {
	listenForPackets()
}
