package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type ChatMessage struct {
	Message string `json:"m"`
	User    string `json:"u"`
	Color   string `json:"c"`
}

type UserBlob struct {
	User  string `json:"u"`
	Color string `json:"c"`
}

type userInfo struct {
	Color string
	User  string
	Loc   string
}

var lastTimestamp int64
var isMsgOutgoing = false
var mu sync.Mutex

var firstSeenHash = map[string][]byte{}     // first hash seen for each username
var mixedHashToUser = map[string]userInfo{} // user name, color, loc based on mixed hash
var seenNotes = map[string]struct{}{}       // cache of note dedups sent by server
var onlineUsers = map[string]userInfo{}     // cache of known users currently online - list of userInfos

func sendInfoPing(dest string) string {
	return string(sendPacket("info", []byte{}, dest))
}

func personalHash() []byte {
	h := sha256.Sum256(append([]byte(*sign), []byte(*user)...))
	return h[:16]
}

func bringOnline(hash []byte, u userInfo) {
	onlineUsers[string(hash)] = u
	app.QueueUpdateDraw(redrawUserView)

	// the len(firstSeenHash) check checks to make sure it doesn't print before any normal messages (user just joined convo)
	if u.User != *user && len(firstSeenHash) != 0 {
		tuiPrint("[slategray]" + u.User + " came online")
	}
}

func bringOffline(hash []byte, u userInfo) {
	delete(onlineUsers, string(hash))
	app.QueueUpdateDraw(redrawUserView)
	if len(firstSeenHash) != 0 {
		tuiPrint("[slategray]" + u.User + " went offline")
	}
}

func processNotes(response MsgRecord) {
	for _, note := range response.PendingNotes {
		if _, seen := seenNotes[note.DedupHash]; seen {
			continue
		}
		seenNotes[note.DedupHash] = struct{}{}

		if len(note.Payload) > 0 {
			u := decryptUserBlob(note.Payload)
			if u == nil {
				continue
			}
			u.Loc = note.Loc
			mixedHashToUser[string(note.Referent)] = *u
			if _, isOnline := onlineUsers[string(note.Referent)]; !isOnline {
				bringOnline(note.Referent, *u)
			}
		} else {
			if u, isOnline := onlineUsers[string(note.Referent)]; isOnline {
				bringOffline(note.Referent, u)
			}
		}
	}
}

// this runs to parse the response from the server returning a ping
func handleResponse(responseBytes []byte) {
	mu.Lock()
	defer mu.Unlock()

	// decrypt the message encrypted by the server using the hash of the password
	responseStr := decryptFromBytes(responseBytes, deobfuscate(obfuscate(passHash(*pass))))
	var serverResponse MsgRecord
	if err := json.Unmarshal(responseStr, &serverResponse); err != nil {
		return
	}

	var lastMessage ChatMessage

	// now decrypt the internal message originally sent by the other client
	msgTextStr := decryptUsingPass(serverResponse.MsgPayload, *pass)
	json.Unmarshal([]byte(msgTextStr), &lastMessage)

	convoIsNotNew := serverResponse.LastMixedHash != nil && lastMessage.User != ""

	// now we have both lastMessage (extends ChatMessage) and serverResponse (extends MsgRecord)
	// (this is where the fun begins)

	if convoIsNotNew {

		// if it's been more than 5 minutes since last message print the time
		if serverResponse.MsgTimestamp >= lastTimestamp+5*60 {
			if len(firstSeenHash) != 0 {
				// print a spacing line if the user did not just enter the convo
				tuiPrint("")
			}
			tuiPrint("[slategray]" + formatTimestamp(serverResponse.MsgTimestamp))
		}

		// add the user data for this message to the lookup map for hash -> user info
		mixedHashToUser[string(serverResponse.LastMixedHash)] = userInfo{lastMessage.Color, lastMessage.User, serverResponse.IpLocation}

		// check if this hash has been seen before; if it hasn't, add it to the lookup map for hash -> first user seen with that hash
		// and also set the user to be online (quietly because it's assuming based only on message recency)
		if _, seen := firstSeenHash[lastMessage.User]; !seen {

			// it gets to here if this latest hash HASN'T been seen before
			firstSeenHash[lastMessage.User] = serverResponse.LastMixedHash

			// if the message was sent in the last 20 seconds OR the message was sent by the user, set the sender to online (quietly)
			if time.Now().Unix()-serverResponse.MsgTimestamp <= 20 || lastMessage.User == *user {
				if _, online := onlineUsers[string(serverResponse.LastMixedHash)]; !online {
					onlineUsers[string(serverResponse.LastMixedHash)] = userInfo{lastMessage.Color, lastMessage.User, serverResponse.IpLocation}
					app.QueueUpdateDraw(redrawUserView)
				}
			}
		}

		// check for any notes from the server
		// any online or offline updates are printed in this function
		processNotes(serverResponse)
	}

	// determine if the message is new using the timestamp provided by the server. if it is new, print it
	// this block is run IF THERE IS A NEW MESSAGE
	if serverResponse.MsgTimestamp != lastTimestamp {
		if convoIsNotNew {
			if first := firstSeenHash[lastMessage.User]; !bytes.Equal(first, serverResponse.LastMixedHash) {
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

	// send passHash + personalHash + encrypted message
	payload := append(append(obfuscate(passHash(*pass)), obfuscate(personalHash())...), encryptToBytes(jsonBytes, []byte(*pass))...)
	responseBytes := sendPacket("msg", payload, *ip)
	if responseBytes != nil {
		handleResponse(responseBytes)
		isMsgOutgoing = false
	}
}

func sendHandshake() {
	ub, _ := json.Marshal(UserBlob{User: *user, Color: *color})
	blob := encryptToBytes(ub, []byte(*pass))
	payload := append(append(obfuscate(passHash(*pass)), obfuscate(personalHash())...), blob...)
	responseBytes := sendPacket("shake", payload, *ip)
	if responseBytes == nil {
		return
	}
	decrypted := decryptFromBytes(responseBytes, deobfuscate(obfuscate(passHash(*pass))))
	type handshakeResponse struct {
		Total int         `json:"t"`
		Users []UserEntry `json:"u"`
	}
	var resp handshakeResponse
	if err := json.Unmarshal(decrypted, &resp); err != nil {
		tuiPrint(err.Error())
	}

	for _, u := range resp.Users {
		info := decryptUserBlob(u.Blob)
		if info == nil {
			continue
		}
		info.Loc = u.Loc
		mixedHashToUser[string(u.Hash)] = *info
		bringOnline(u.Hash, *info)
	}
	if resp.Total > len(resp.Users) {
		tuiPrint(fmt.Sprintf("[slategray]%d users not loaded", resp.Total-len(resp.Users)))
	}
}

func runClientListener() {
	// send a poll with passHash + personalHash + dummy bytes a few times a second with randomish delay
	for {
		const chars = "abcdefghij0123456789"
		salt := []byte(chars)
		pollPayload := append(append(obfuscate(passHash(*pass)), obfuscate(personalHash())...), salt...)
		responseBytes := sendPacket("poll", pollPayload, *ip)
		if responseBytes != nil {
			handleResponse(responseBytes)
		}
		time.Sleep(time.Duration(160+rand.IntN(161)) * time.Millisecond)
	}
}

func formatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	return fmt.Sprintf("%d. %s %d, %02d:%02d", t.Day(), t.Month().String(), t.Year(), t.Hour(), t.Minute())
}
