package main

import (
	"encoding/json"
	"time"
)

type ChatMessage struct {
	Message string `json:"msg"`
	User    string `json:"user"`
	Color   string `json:"color"`
}

var lastTimestamp int

func runClientSender(msg string) {
	msgJson := ChatMessage{Message: msg, User: *user, Color: *color}
	jsonBytes, _ := json.Marshal(msgJson)
	hash := passHash(*pass)
	sendBytes(append(hash, encryptToBytes(jsonBytes, []byte(*pass))...), *ip)
}

func runClientListener() {
	for {
		passwordHashBytes := passHash(*pass)
		responseBytes := sendBytes(passwordHashBytes, *ip)
		responseStr := decryptFromBytes(responseBytes, passwordHashBytes)
		var response MsgRecord
		err := json.Unmarshal(responseStr, &response)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		var incomingMsgJson ChatMessage
		msgTextStr := decryptUsingPass(response.LastMsgEncrypted, *pass)
		json.Unmarshal([]byte(msgTextStr), &incomingMsgJson)
		if response.LastMsgTimestamp != lastTimestamp {
			if incomingMsgJson.Message == "" && incomingMsgJson.User == "" {
				tuiPrint("Chat begins here")
			} else {
				tuiPrint("[" + incomingMsgJson.Color + "]" + incomingMsgJson.User + "[white]: " + incomingMsgJson.Message)
			}
			lastTimestamp = response.LastMsgTimestamp
		}
		time.Sleep(1 * time.Second)
	}
}

func runClient() {
	initTUI(runClientSender)
	go runClientListener()
	runTUI()
}
