package main

import (
	"flag"
)

var (
	server       = flag.Bool("server", false, "run as a server")
	send         = flag.Bool("send", false, "run as a sending term")
	reauthKernel = flag.Bool("reauthKernel", false, "give authorization back to kernel to handle pings")
	user         = flag.String("user", "guest", "user to chat as")
	pass         = flag.String("pass", "", "shared password to chat using")
	ip           = flag.String("ip", "127.0.0.1", "server ip to connect to")
)

func main() {
	flag.Parse()
	if *reauthKernel {
		enableKernelReplies(true)
	} else if *server {
		enableKernelReplies(false)
		runServer()
	} else if *send {
		runClientSender()
	} else {
		runClientListener()
	}
}
