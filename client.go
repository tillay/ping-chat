package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	Loc   string
}

var firstSeenHash = map[string]string{}     // first hash seen for each username
var mixedHashToUser = map[string]userInfo{} // user name, color, loc based on mixed hash
var seenNotes = map[string]bool{}           // cache of note dedups sent by server (todo: make this a normal list)
var onlineUsers = map[string]userInfo{}     // cache of known users currently online - list of userInfos

func personalHash() []byte {
	h := sha256.Sum256(append(append([]byte(*sign), []byte(*user)...), []byte(*color)...))
	return h[:16]
}

func redrawUserView() {
	usersView.SetText("")
	for _, u := range onlineUsers {
		fmt.Fprintf(usersView, "[%s]%s[white]\n%s\n\n", u.Color, u.User, u.Loc)
	}
}

func bringOnline(hash string, u userInfo) {
	onlineUsers[hash] = u
	app.QueueUpdateDraw(redrawUserView)
	if u.User != *user {
		tuiPrint("[slategray]" + u.User + " came online")
	}
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
		if rawNote, _ := hex.DecodeString(note.Referent); len(rawNote) != 8 {
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

// this runs to parse the response from the server returning a ping
func handleResponse(responseBytes []byte) {
	mu.Lock()
	defer mu.Unlock()
	passwordHashBytes := passHash(*pass)
	// decrypt the message encrypted by the server using the hash of the password
	responseStr := decryptFromBytes(responseBytes, passwordHashBytes)
	var serverResponse MsgRecord
	if err := json.Unmarshal(responseStr, &serverResponse); err != nil {
		return
	}

	var lastMessage ChatMessage

	// now decrypt the internal message originally sent by the other client
	msgTextStr := decryptUsingPass(serverResponse.MsgPayload, *pass)
	json.Unmarshal([]byte(msgTextStr), &lastMessage)

	convoIsNotNew := serverResponse.LastMixedHash != "" && lastMessage.User != ""

	// now we have both lastMessage (extends ChatMessage) and serverResponse (extends MsgRecord)
	// (this is where the fun begins)

	// this entire block is ONLY RUN AFTER first msg in convo sent
	if convoIsNotNew {

		// if it's been more than 5 minutes since last message print the time
		if serverResponse.MsgTimestamp >= lastTimestamp+5*60 {
			if convoIsNotNew {
				tuiPrint("")
			}
			tuiPrint("[slategray]" + formatTimestamp(serverResponse.MsgTimestamp))
		}

		// add the user data for this message to the lookup map for hash -> user info
		mixedHashToUser[serverResponse.LastMixedHash] = userInfo{lastMessage.Color, lastMessage.User, serverResponse.IpLocation}

		// check if this hash has been seen before; if it hasn't, add it to the lookup map for hash -> first user seen with that hash
		// and also set it to be online (quietly because it's assuming it is only because message sent recently)
		if _, seen := firstSeenHash[lastMessage.User]; !seen {
			// this runs if this hash HASN'T been seen before
			firstSeenHash[lastMessage.User] = serverResponse.LastMixedHash

			// if it hasn't seen this hash before but the message was sent in the last 20 seconds, set the user to online (quietly)
			if time.Now().Unix()-serverResponse.MsgTimestamp <= 20 {
				if _, online := onlineUsers[serverResponse.LastMixedHash]; !online {
					onlineUsers[serverResponse.LastMixedHash] = userInfo{lastMessage.Color, lastMessage.User, serverResponse.IpLocation}
					app.QueueUpdateDraw(redrawUserView)
				}
			}
		}

		// check for any notes from the server
		// any online or offline updates are printed in this function
		processNotes(serverResponse)
	}

	// determine if the message is new using the timestamp provided by the server. if it is new, print it
	// this block is only run IF THERE IS A NEW MESSAGE and CHAT IS NOT NEW
	if serverResponse.MsgTimestamp != lastTimestamp {
		if convoIsNotNew {
			if first := firstSeenHash[lastMessage.User]; first != serverResponse.LastMixedHash {
				tuiPrint("[yellow]" + lastMessage.User + " has invalid signature")
			}
			tuiPrint("[-:-:-][" + lastMessage.Color + "]" + lastMessage.User + "[-:-:-][white]: " + lastMessage.Message)
		}
		lastTimestamp = serverResponse.MsgTimestamp
	}
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
