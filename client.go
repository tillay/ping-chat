package main

import (
	"encoding/json"
	"slices"
	"time"
)

type ChatMessage struct {
	Message string `json:"msg"`
	User    string `json:"user"`
	Color   string `json:"color"`
}

var lastTimestamp int
var users []string

// this runs to parse the response from the server returning a ping
func handleResponse(responseBytes []byte) {
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

	// determine if the message is new using the timestamp provided by the server. if it is new, print it
	if response.MsgTimestamp != lastTimestamp {
		if incomingMsgJson.Message == "" && incomingMsgJson.User == "" {
			tuiPrint("Chat begins here")
		} else {
			compressedUserData := incomingMsgJson.Color + incomingMsgJson.User + response.MsgIp + response.MsgIp
			if !slices.Contains(users, compressedUserData) {
				users = append(users, compressedUserData)
				sideViewPrint("[" + incomingMsgJson.Color + "]" + incomingMsgJson.User + "[white]: " + response.IpLocation)
			}
			tuiPrint("[" + incomingMsgJson.Color + "]" + incomingMsgJson.User + "[white]: " + incomingMsgJson.Message)
		}
		lastTimestamp = response.MsgTimestamp
	}
}

func runClientSender(msg string) {
	msgJson := ChatMessage{Message: msg, User: *user, Color: *color}
	jsonBytes, _ := json.Marshal(msgJson)
	hash := passHash(*pass)
	// send the hash of the password stuck to the start of the encrypted data to the server.
	// the server will use this hash to tell conversations apart and also encrypt for the other clients
	responseBytes := sendBytes(append(hash, encryptToBytes(jsonBytes, []byte(*pass))...), *ip)
	if responseBytes != nil {
		handleResponse(responseBytes)
	}
}

func runClientListener() {
	// every second send a ping with contents of the hash of the password
	// server always responds with the most recent message of the conversation
	for {
		passwordHashBytes := passHash(*pass)
		responseBytes := sendBytes(passwordHashBytes, *ip)
		if responseBytes != nil {
			handleResponse(responseBytes)
		}
		time.Sleep(1 * time.Second)
	}
}

func runClient() {
	initTUI(runClientSender)
	go runClientListener()
	runTUI()
}
