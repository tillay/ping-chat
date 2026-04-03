package main

import (
	"flag"
)

var (
	server       = flag.Bool("server", false, "run as a server")
	reauthKernel = flag.Bool("reauthKernel", false, "give authorization back to kernel to handle pings")
)
//unused main method, either
//func main() {
//	flag.Parse()
//	if *reauthKernel {
//		enableKernelReplies(true)
//	} else if *server {
//		enableKernelReplies(false)
//		listenForPackets()
//	} else {
//		fmt.Println(sendString(os.Args[1], "91.98.131.91"))
//	}
//}
