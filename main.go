package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	server = flag.Bool("server", false, "run instance as server for message coordinating and forwarding")
	salt   = flag.String("salt", "", "signature for server to scramble user hashes with in order to prevent impersonation")
	force  = flag.Bool("login", false, "force re-creation of config file and overwrite any existing there")
	path   = flag.String("path", filepath.Join(os.Getenv("HOME"), ".config", "pingchat"), "path to config file")
)

var User Config

func main() {
	flag.Parse()
	if *server {
		if *salt == "" {
			fmt.Println("Error: you must specify a signature with -salt to prevent impersonation")
			return
		}
		enableKernelReplies(false)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sig
			enableKernelReplies(true)
			os.Exit(0)
		}()
		listenForPackets()
	} else {
		User = loadConfig(*force)
		saveConfig(User)
		initTUI(runClientSender)
		go func() { sendHandshake(); runClientListener() }()
		runTUI()
	}
}
