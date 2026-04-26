package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	server = flag.Bool("server", false, "run as a server")
	user   = flag.String("user", "", "user to chat as")
	pass   = flag.String("pass", "", "shared password to for chat (everyone in gc uses this same pass)")
	ip     = flag.String("ip", "", "server ip to connect to")
	color  = flag.String("color", "#ffffff", "color to set your name")
	sign   = flag.String("sign", "", "signature to sign messages with (do not share)")
)

var (
	magicPollBytes = genMagicBytes("poll")
	magicMsgBytes  = genMagicBytes("msg")
	magicHsBytes   = genMagicBytes("shake")
	magicInfoBytes = genMagicBytes("info")
)

func main() {
	flag.Parse()
	if *server {
		enableKernelReplies(false)
		listenForPackets()
	} else if *sign != "" && *ip != "" && *user != "" && *pass != "" {
		initTUI(runClientSender)
		go func() { sendHandshake(); runClientListener() }()
		runTUI()
	} else {
		fmt.Println("------\nYou need to specify server ip, user, pass, and sign\n------")
		flag.Usage()
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		enableKernelReplies(true)
		os.Exit(0)
	}()
}
