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

func runClient() {//takes message and password, encrypts, decrypts, checks hash
	var pass string//please note servers will use main at main.go
	var message string

	
	fmt.Println("Input password")
	fmt.Scan(&pass)
	fmt.Println("Input message")
	fmt.Scan(&message)

	encrypted := encryptUsingPass(message, pass)
	fmt.Println(encrypted)
	fmt.Println(decryptUsingPass(encrypted, pass))
	fmt.Println(extractHash(encrypted))
}

func runClientSender() {
	for {
		fmt.Printf("> ")
		msg := ""
		fmt.Scanln(&msg)
		msgJson := ChatMessage{
			Message: msg,
			User:    *user,
		}
		jsonBytes, _ := json.Marshal(msgJson)
		responseBytes := sendBytes(encryptUsingPass(string(jsonBytes), *pass), *ip)
		decryptUsingHash(responseBytes, passHash(*pass))
	}
}

func runClientListener() {
	for {
		passwordHashBytes := passHash(*pass)
		responseBytes := sendBytes(passwordHashBytes, *ip)
		responseStr := decryptUsingHash(responseBytes, passHash(*pass))

		fmt.Println("Decrypted response:", responseStr)
		var response MsgRecord
		err := json.Unmarshal(responseStr, &response)
		if err != nil {
			fmt.Println(response)
			fmt.Println("Error decoding response:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var incomingMsgJson ChatMessage
		msgTextStr := decryptUsingPass(response.LastMsgEncrypted, *pass)
		fmt.Println("decrypted message: " + msgTextStr)
		json.Unmarshal([]byte(msgTextStr), &incomingMsgJson)

		if response.LastMsgTimestamp != lastTimestamp {
			fmt.Println(incomingMsgJson.User + ": " + incomingMsgJson.Message)
			lastTimestamp = response.LastMsgTimestamp
		}
		time.Sleep(1 * time.Second)
	}
}
