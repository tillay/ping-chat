package main

import (
	"flag"
	"fmt"
)

var (
	server       = flag.Bool("server", false, "run as a server")
	reauthKernel = flag.Bool("fix", false, "give authorization back to kernel to handle pings")
	user         = flag.String("user", "guest", "user to chat as")
	pass         = flag.String("pass", "", "shared password to for chat (everyone in gc uses this same pass)")
	ip           = flag.String("ip", "", "server ip to connect to")
	color        = flag.String("color", "#ffffff", "color to set your name")
	info         = flag.Bool("info", false, "send a test ping")
	sign         = flag.String("sign", "", "signature to sign messages with (do not share)")
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
	} else if *sign != "" && *ip != "" && *user != "" && *pass != "" {
		runClient()
	} else {
		fmt.Println("You need to specify server ip, user, pass, and sign")
		flag.Usage()
	}
}
