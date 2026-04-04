package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type ChatMessage struct {
	Message string `json:"msg"`
	User    string `json:"user"`
}

var lastTimestamp int

func runClientSender() {
	for {
		fmt.Printf("> ")
		msg := ""
		fmt.Scanln(&msg)
		msgJson := ChatMessage{Message: msg, User: *user}
		jsonBytes, _ := json.Marshal(msgJson)
		hash := passHash(*pass)
		sendBytes(append(hash, encryptToBytes(jsonBytes, []byte(*pass))...), *ip)
	}
}

func runClientListener() {
	for {
		passwordHashBytes := passHash(*pass)
		responseBytes := sendBytes(passwordHashBytes, *ip)
		responseStr := decryptFromBytes(responseBytes, passwordHashBytes)
		var response MsgRecord
		err := json.Unmarshal(responseStr, &response)

		if err != nil {
			fmt.Println("Error decoding response:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var incomingMsgJson ChatMessage
		msgTextStr := decryptUsingPass(response.LastMsgEncrypted, *pass)

		json.Unmarshal([]byte(msgTextStr), &incomingMsgJson)

		if response.LastMsgTimestamp != lastTimestamp {
			fmt.Println(incomingMsgJson.User + ": " + incomingMsgJson.Message)
			lastTimestamp = response.LastMsgTimestamp

		}
		time.Sleep(1 * time.Second)
	}
}
