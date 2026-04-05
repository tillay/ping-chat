package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func makeChecksum(b []byte) uint16 {
	s := uint32(0)
	for i := 0; i < len(b)-1; i += 2 {
		s += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if len(b)&1 != 0 {
		s += uint32(b[len(b)-1]) << 8
	}
	s = s>>16 + s&0xffff
	return ^uint16(s + s>>16)
}

func enableKernelReplies(val bool) {
	option := "0"
	if !val {
		option = "1" //1 makes system ignore incoming pings
	}
	err := os.WriteFile("/proc/sys/net/ipv4/icmp_echo_ignore_all", []byte(option), 0644)
	processErr(err)
}

func sendBytes(data []byte, dest string) []byte {
	buf := make([]byte, len(data)+11)
	buf[0] = 8
	buf[9], buf[10] = 0x4F, 0x4B
	copy(buf[11:], data)
	s := makeChecksum(buf)
	buf[2], buf[3] = byte(s>>8), byte(s)

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	dst, _ := net.ResolveIPAddr("ip4", dest)
	c.WriteTo(buf, dst)//Send packets

	recv := make([]byte, 1500)
	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		n, addr, err := c.ReadFrom(recv)
		if err != nil {
			fmt.Println("Error: ", err.Error())
			return nil
		}
		if n < 11 || recv[0] != 0 || addr.String() != dest || recv[9] != 0x4F || recv[10] != 0x4B {
			continue
		}
		return recv[11:n]
	}
}

func listenForPackets() {
	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	buf := make([]byte, 1500)
	for {
		n, addr, _ := c.ReadFrom(buf)
		if n < 8 || buf[0] != 8 {
			continue
		}
		var msg []byte
		if n >= 11 && buf[9] == 0x4F && buf[10] == 0x4B {
			payload := make([]byte, n-11)
			copy(payload, buf[11:n])
			fmt.Printf("%s: (%d bytes)\n", addr, len(payload))
			msg = sendReply(addr.String(), payload)
			reply := make([]byte, 11+len(msg))
			reply[0] = 0
			reply[9], reply[10] = 0x4F, 0x4B
			copy(reply[4:8], buf[4:8])
			copy(reply[11:], msg)
			reply[2], reply[3] = 0, 0
			s := makeChecksum(reply)
			reply[2], reply[3] = byte(s>>8), byte(s)
			c.WriteTo(reply, addr)
		} else {
			reply := make([]byte, n)
			reply[0] = 0
			copy(reply[4:], buf[4:n])
			reply[2], reply[3] = 0, 0
			s := makeChecksum(reply)
			reply[2], reply[3] = byte(s>>8), byte(s)
			c.WriteTo(reply, addr)
		}
	}
}

func sendString(data string, dest string) string {
	return string(sendBytes([]byte(data), dest))
}

func processErr(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			fmt.Println("Error: this program needs root")
			os.Exit(1)
		}
		panic(err)
	}
}
