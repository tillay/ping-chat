package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type ChatMessage struct {
	Message string `json:"msg"`
	User    string `json:"user"`
	Color   string `json:"color"`
}

var lastTimestamp int64
var isMsgOutgoing = false
var mu sync.Mutex

type userInfo struct {
	Color string
	User  string
}

var firstSeenHash = map[string]string{}     // first hash seen for each username
var mixedHashToUser = map[string]userInfo{} // username and color based on mixed hash
var seenNotes = map[string]bool{}           // cache of note dedups sent by server
var onlineUsers = map[string]userInfo{}     // cache of known users currently online

func personalHash() []byte {
	h := sha256.Sum256(append(append([]byte(*sign), []byte(*user)...), []byte(*color)...))
	return h[:16]
}

func redrawUserView() {
	usersView.SetText("")
	for _, u := range onlineUsers {
		fmt.Fprintf(usersView, "[%s]%s\n", u.Color, u.User)
	}
}

func bringOnline(hash string, u userInfo) {
	onlineUsers[hash] = u
	app.QueueUpdateDraw(redrawUserView)
	tuiPrint("[slategray]" + u.User + " came online")
}

func bringOffline(hash string, u userInfo) {
	delete(onlineUsers, hash)
	app.QueueUpdateDraw(redrawUserView)
	tuiPrint("[slategray]" + u.User + " went offline")
}

func processNotes(response MsgRecord) {
	for _, note := range response.PendingNotes {
		if seenNotes[note.DedupHash] {
			continue
		}
		if note.Referent == "info" {
			seenNotes[note.DedupHash] = true
			tuiPrint(note.Info)
			continue
		}
		u, known := mixedHashToUser[note.Referent]
		if !known {
			continue
		}
		seenNotes[note.DedupHash] = true
		_, isOnline := onlineUsers[note.Referent]
		switch note.Info {
		case "online":
			if !isOnline {
				bringOnline(note.Referent, u)
			}
		case "offline":
			if isOnline {
				bringOffline(note.Referent, u)
			}
		}
	}
}

func senderOnlineNote(response MsgRecord) (found bool) {
	for _, n := range response.PendingNotes {
		if n.Referent != response.LastMixedHash || n.Info != "online" || seenNotes[n.DedupHash] {
			continue
		}
		if _, isOnline := onlineUsers[n.Referent]; !isOnline {
			bringOnline(n.Referent, mixedHashToUser[response.LastMixedHash])
		}
		seenNotes[n.DedupHash] = true
		return true
	}
	return false
}

// this runs to parse the response from the server returning a ping
func handleResponse(responseBytes []byte) {
	mu.Lock()
	defer mu.Unlock()
	passwordHashBytes := passHash(*pass)
	// decrypt the message encrypted by the server using the hash of the password
	responseStr := decryptFromBytes(responseBytes, passwordHashBytes)
	var response MsgRecord
	if err := json.Unmarshal(responseStr, &response); err != nil {
		return
	}

	var incomingMsgJson ChatMessage
	// now decrypt the internal message originally sent by the other client
	msgTextStr := decryptUsingPass(response.MsgPayload, *pass)
	json.Unmarshal([]byte(msgTextStr), &incomingMsgJson)

	if response.LastMixedHash != "" && incomingMsgJson.User != "" {
		mixedHashToUser[response.LastMixedHash] = userInfo{incomingMsgJson.Color, incomingMsgJson.User}
		if _, seen := firstSeenHash[incomingMsgJson.User]; !seen {
			firstSeenHash[incomingMsgJson.User] = response.LastMixedHash
		}
	}

	// determine if the message is new using the timestamp provided by the server. if it is new, print it
	if response.MsgTimestamp != lastTimestamp {
		if response.MsgTimestamp >= lastTimestamp+5*60 {
			tuiPrint("")
			tuiPrint("[slategray]" + formatTimestamp(response.MsgTimestamp))
		}
		if incomingMsgJson.Message != "" || incomingMsgJson.User != "" {
			if first := firstSeenHash[incomingMsgJson.User]; first != "" && first != response.LastMixedHash {
				tuiPrint("[yellow]" + incomingMsgJson.User + " has invalid signature")
			}
			hasOnlineNote := senderOnlineNote(response)
			if !hasOnlineNote && time.Now().Unix()-response.MsgTimestamp <= 20 {
				if _, online := onlineUsers[response.LastMixedHash]; !online {
					onlineUsers[response.LastMixedHash] = userInfo{incomingMsgJson.Color, incomingMsgJson.User}
					app.QueueUpdateDraw(redrawUserView)
				}
			}
			tuiPrint("[-:-:-][" + incomingMsgJson.Color + "]" + incomingMsgJson.User + "[-:-:-][white]: " + incomingMsgJson.Message)
		}
		lastTimestamp = response.MsgTimestamp
	}

	processNotes(response)
}

func runClientSender(msg string) {
	msgJson := ChatMessage{Message: msg, User: *user, Color: *color}
	jsonBytes, _ := json.Marshal(msgJson)
	hash := passHash(*pass)
	ph := personalHash()
	// send passHash + personalHash + encrypted message
	payload := append(append(hash, ph...), encryptToBytes(jsonBytes, []byte(*pass))...)
	responseBytes := sendBytes(payload, *ip)
	if responseBytes != nil {
		handleResponse(responseBytes)
		isMsgOutgoing = false
	}
}

func runClientListener() {
	// every 200ms send a poll with passHash + personalHash
	for {
		poll := append(passHash(*pass), personalHash()...)
		responseBytes := sendBytes(poll, *ip)
		if responseBytes != nil {
			handleResponse(responseBytes)
		}
		time.Sleep(time.Second / 5)
	}
}

func formatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	return fmt.Sprintf("%d. %s %d, %02d:%02d", t.Day(), t.Month().String(), t.Year(), t.Hour(), t.Minute())
}

func runClient() {
	initTUI(runClientSender)
	go runClientListener()
	runTUI()
}
