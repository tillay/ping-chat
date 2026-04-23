package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type NoteEntry struct {
	DedupHash string
	Referent  string
	Info      string
	CreatedAt int64 `json:"-"`
}

type MsgRecord struct {
	MsgIp         string `json:"-"`
	MsgPayload    []byte
	MsgTimestamp  int64
	IpLocation    string
	LastMixedHash string
	OnlineHashes  map[string]int64 `json:"-"`
	PendingNotes  []NoteEntry
}

var store = map[string]MsgRecord{}
var locationCache = map[string]string{}

func mixedHash(ip string, personalHash []byte) string {
	h := sha256.Sum256(append([]byte(ip), personalHash...))
	return hex.EncodeToString(h[:8])
}

func dedupHash() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func purgeNotesForHash(record *MsgRecord, hash string) {
	filtered := make([]NoteEntry, 0, len(record.PendingNotes))
	for _, n := range record.PendingNotes {
		if n.Referent != hash {
			filtered = append(filtered, n)
		}
	}
	record.PendingNotes = filtered
}

func sweepRecord(record *MsgRecord, senderMixed string) {
	now := time.Now().Unix()
	for h, ts := range record.OnlineHashes {
		if now-ts > 2 {
			delete(record.OnlineHashes, h)
			purgeNotesForHash(record, h)
			record.PendingNotes = append(record.PendingNotes, NoteEntry{
				DedupHash: dedupHash(),
				Referent:  h,
				Info:      "offline",
				CreatedAt: now,
			})
		}
	}

	filtered := make([]NoteEntry, 0, len(record.PendingNotes))
	for _, n := range record.PendingNotes {
		if now-n.CreatedAt <= 20 {
			filtered = append(filtered, n)
		}
	}
	record.PendingNotes = filtered

	if _, exists := record.OnlineHashes[senderMixed]; !exists {
		purgeNotesForHash(record, senderMixed)
		record.PendingNotes = append(record.PendingNotes, NoteEntry{
			DedupHash: dedupHash(),
			Referent:  senderMixed,
			Info:      "online",
			CreatedAt: now,
		})
	}
	record.OnlineHashes[senderMixed] = now
}

func getOrCreateRecord(key []byte, ip string) MsgRecord {
	record, exists := store[string(key)]
	if !exists {
		record = MsgRecord{
			MsgIp:        ip,
			MsgTimestamp: time.Now().Unix(),
			IpLocation:   getIpLocation(ip),
			OnlineHashes: map[string]int64{},
			PendingNotes: []NoteEntry{},
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

func sendReply(ip string, incomingPayload []byte) []byte {
	key := extractHash(incomingPayload)
	record := getOrCreateRecord(key, ip)

	if len(incomingPayload) < 48 {
		return incomingPayload
	}

	personalHash := incomingPayload[32:48]
	senderMixed := mixedHash(ip, personalHash)

	sweepRecord(&record, senderMixed)

	if len(incomingPayload) > 48 {
		record.MsgPayload = incomingPayload[48:]
		record.MsgIp = ip
		record.IpLocation = getIpLocation(ip)
		record.MsgTimestamp = time.Now().Unix()
		record.LastMixedHash = senderMixed
	}

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
