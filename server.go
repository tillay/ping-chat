package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MsgRecord struct {
	MsgIp        string `json:"-"`
	MsgPayload   []byte
	MsgTimestamp int64
	IpLocation   string
	IpTimestamps map[string]int64 `json:"-"`
	Note         string
}

var store = map[string]MsgRecord{}
var locationCache = map[string]string{}

// todo: hash ip and send it to client
// then include in notes that that hash has either come online or gone offline
// and then client can somehow get the user(s?) with that hash and print that said user has changed online status
// (but how to handle it if client doesnt have a hash for that user?)
func getOrCreateRecord(key []byte, ip string) MsgRecord {
	record, exists := store[string(key)]
	if !exists {
		record = MsgRecord{
			MsgIp:        ip,
			MsgTimestamp: time.Now().Unix(),
			IpLocation:   getIpLocation(ip),
			IpTimestamps: map[string]int64{},
		}
	}
	return record
}

func saveAndReply(record MsgRecord, key []byte) []byte {
	store[string(key)] = record
	recordJson, _ := json.Marshal(record)
	fmt.Println(string(recordJson))
	return encryptToBytes(recordJson, key)
}

// reply to any incoming pings
// gets incoming message ip and encrypted bytes from the client, returns what to reply with
func sendReply(ip string, incomingPayload []byte) []byte {
	key := extractHash(incomingPayload)
	record := getOrCreateRecord(key, ip)

	if len(incomingPayload) == 32 {
		record.IpTimestamps[ip] = time.Now().Unix()
	} else {
		record.MsgPayload = incomingPayload[32:]
		record.MsgIp = ip
		record.IpLocation = getIpLocation(ip)
		record.MsgTimestamp = time.Now().Unix()
	}

	record.IpTimestamps[ip] = time.Now().Unix()
	return saveAndReply(record, key)
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
