package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

type NoteEntry struct {
	DedupHash string `json:"d"`
	Referent  string `json:"r"`
	Payload   []byte `json:"b,omitempty"`
	Loc       string `json:"l,omitempty"`
	CreatedAt int64  `json:"-"`
}

type UserEntry struct {
	Hash string `json:"x"`
	Blob []byte `json:"b"`
	Loc  string `json:"l,omitempty"`
	Seen int64  `json:"-"`
}

type MsgRecord struct {
	MsgIp         string      `json:"-"`
	MsgPayload    []byte      `json:"p"`
	MsgTimestamp  int64       `json:"t"`
	IpLocation    string      `json:"l"`
	LastMixedHash string      `json:"h"`
	KnownUsers    []UserEntry `json:"-"`
	PendingNotes  []NoteEntry `json:"n"`
}

var store = map[string]MsgRecord{}
var locationCache = map[string]map[string]string{}

func mixedHash(ip string, personalHash []byte) string {
	h := sha256.Sum256(append([]byte(ip), personalHash...))
	return hex.EncodeToString(h[:8])
}

func dedupHash() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func findUser(record *MsgRecord, hash string) *UserEntry {
	for i := range record.KnownUsers {
		if record.KnownUsers[i].Hash == hash {
			return &record.KnownUsers[i]
		}
	}
	return nil
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

	for i := range record.KnownUsers {
		u := &record.KnownUsers[i]
		if now-u.Seen > 2 && now-u.Seen <= 20 {
			purgeNotesForHash(record, u.Hash)
			record.PendingNotes = append(record.PendingNotes, NoteEntry{
				DedupHash: dedupHash(),
				Referent:  u.Hash,
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

	u := findUser(record, senderMixed)
	if u == nil || now-u.Seen > 2 {
		purgeNotesForHash(record, senderMixed)
		var blob []byte
		var loc string
		if u != nil {
			blob = u.Blob
			loc = u.Loc
		}
		record.PendingNotes = append(record.PendingNotes, NoteEntry{
			DedupHash: dedupHash(),
			Referent:  senderMixed,
			Payload:   blob,
			Loc:       loc,
			CreatedAt: now,
		})
	}
	if u != nil {
		u.Seen = now
	}
}

func getOrCreateRecord(key []byte, ip string) MsgRecord {
	record, exists := store[string(key)]
	if !exists {
		record = MsgRecord{
			MsgIp:        ip,
			MsgTimestamp: time.Now().Unix(),
			IpLocation:   getIpLocation(ip),
			KnownUsers:   []UserEntry{},
			PendingNotes: []NoteEntry{},
		}
	}
	return record
}

func saveAndReply(record MsgRecord, key []byte) []byte {
	store[string(key)] = record
	recordJson, _ := json.Marshal(record)
	fmt.Println("Replying " + string(recordJson))
	return encryptToBytes(recordJson, key)
}

func handlePoll(addr net.Addr, payload []byte) []byte {
	if len(payload) < 32 {
		return payload
	}
	key := extractHash(payload)
	record := getOrCreateRecord(key, addr.String())
	senderMixed := mixedHash(addr.String(), payload[16:32])
	sweepRecord(&record, senderMixed)
	fmt.Printf("Incoming polling ping from %s in %x (%d bytes)\n", addr.String(), key, len(payload))
	return saveAndReply(record, key)
}

func handleMessage(addr net.Addr, payload []byte) []byte {
	if len(payload) < 32 {
		return payload
	}
	key := extractHash(payload)
	record := getOrCreateRecord(key, addr.String())
	senderMixed := mixedHash(addr.String(), payload[16:32])
	fmt.Printf("Incoming message from %s in %x (%d bytes)\n", addr.String(), key, len(payload))
	sweepRecord(&record, senderMixed)
	record.MsgPayload = payload[32:]
	record.MsgIp = addr.String()
	record.IpLocation = getIpLocation(addr.String())
	record.MsgTimestamp = time.Now().Unix()
	record.LastMixedHash = senderMixed
	return saveAndReply(record, key)
}

func handleHandshake(addr net.Addr, payload []byte) []byte {
	if len(payload) < 32 {
		return payload
	}
	key := extractHash(payload)
	record := getOrCreateRecord(key, addr.String())
	senderMixed := mixedHash(addr.String(), payload[16:32])
	userBlob := payload[32:]
	fmt.Printf("Incoming handshake from %s in %x (%d bytes)\n", addr.String(), key, len(payload))
	loc := getIpLocation(addr.String())
	u := findUser(&record, senderMixed)
	if u == nil {
		// Seen left at zero so sweepRecord sees the user as new and emits an online note
		record.KnownUsers = append(record.KnownUsers, UserEntry{
			Hash: senderMixed,
			Blob: userBlob,
			Loc:  loc,
		})
	} else {
		u.Blob = userBlob
		u.Loc = loc
		// Seen NOT updated here; sweepRecord will emit an online note if they were away
	}
	sweepRecord(&record, senderMixed)
	store[string(key)] = record

	active := make([]UserEntry, 0)
	for _, knownU := range record.KnownUsers {
		if time.Now().Unix()-knownU.Seen <= 2 {
			active = append(active, knownU)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].Seen > active[j].Seen
	})
	totalActive := len(active)
	if totalActive > 12 {
		active = active[:12]
	}

	type handshakeResponse struct {
		Total int         `json:"t"`
		Users []UserEntry `json:"u"`
	}
	resp, _ := json.Marshal(handshakeResponse{
		Total: totalActive,
		Users: active,
	})
	return encryptToBytes(resp, key)
}

func getIpData(ip string) map[string]string {
	if loc, ok := locationCache[ip]; ok {
		return loc
	}
	resp, err := http.Get("https://ipinfo.io/" + ip + "/json")
	if err != nil {
		return map[string]string{}
	}
	defer resp.Body.Close()
	var data map[string]string
	json.NewDecoder(resp.Body).Decode(&data)
	locationCache[ip] = data
	return data
}

func getIpLocation(ip string) string {
	ipData := getIpData(ip)
	return ipData["city"] + ", " + ipData["country"]
}

func buildInfoReply(info map[string]string) []byte {
	org := info["org"]
	if idx := strings.Index(org, " "); idx != -1 {
		org = org[idx+1:]
	}
	return []byte(info["ip"] + " | " + info["city"] + ", " + info["region"] + ", " + info["country"] + " | " + org)
}

func handleInfoPing(addr net.Addr, payload []byte) []byte {
	fmt.Printf("Incoming handshake from %s (%d bytes)\n", addr.String(), len(payload))
	return buildInfoReply(getIpData(addr.String()))
}
