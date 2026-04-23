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
	test         = flag.Bool("test", false, "send a test ping")
	sign         = flag.String("sign", "", "signature to sign messages with (do not share)")
)

func main() {
	flag.Parse()
	if *reauthKernel {
		enableKernelReplies(true)
	} else if *test {
		if *ip != "" {
			fmt.Println(sendInfoPing(*ip))
		} else {
			fmt.Println("you need to specify ip to ping")
		}
	} else if *server {
		enableKernelReplies(false)
		listenForPackets()
	} else if *sign != "" && *ip != "" && *user != "" && *pass != "" {
		runClient()
	} else {
		fmt.Println("------\nYou need to specify server ip, user, pass, and sign\n------")
		flag.Usage()
	}
}
