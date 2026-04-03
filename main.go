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

func main() { //takes input and redirects to run either client or server
	mode := flag.String("mode", "client", "inputs server or client")
	flag.Parse()

	if *mode == "server"{
		runServer()
	} else {
		runClient()
	}
}
