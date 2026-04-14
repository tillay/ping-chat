package main

import (
	"flag"
	"fmt"
)

var (
	server       = flag.Bool("server", false, "run as a server")
	reauthKernel = flag.Bool("reauthKernel", false, "give authorization back to kernel to handle pings")
	user         = flag.String("user", "guest", "user to chat as")
	pass         = flag.String("pass", "", "shared password to chat using")
	ip           = flag.String("ip", "127.0.0.1", "server ip to connect to")
	color        = flag.String("color", "#ffffff", "color output")
	info         = flag.Bool("info", false, "send an info ping and print result")
)

func main() {
	flag.Parse()
	if *reauthKernel {
		enableKernelReplies(true)
	} else if *info {
		fmt.Println(sendInfoPing(*ip))
	} else if *server {
		enableKernelReplies(false)
		runServer()
	} else {
		runClient()
	}
}
