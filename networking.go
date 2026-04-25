package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func genMagicBytes(sauce string) []byte {
	h := sha256.Sum256([]byte(sauce))
	return h[:4]
}

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
		option = "1"
	}
	err := os.WriteFile("/proc/sys/net/ipv4/icmp_echo_ignore_all", []byte(option), 0644)
	processErr(err)
	fmt.Println("successfully set kernel icmp replies to " + strconv.FormatBool(val))
}

func buildPacket(magic []byte, data []byte) []byte {
	buf := make([]byte, 12+len(data))
	buf[0] = 8
	copy(buf[8:12], magic)
	copy(buf[12:], data)
	s := makeChecksum(buf)
	buf[2], buf[3] = byte(s>>8), byte(s)
	return buf
}

func sendPacket(magic []byte, data []byte, dest string) []byte {
	buf := buildPacket(magic, data)

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	dst, _ := net.ResolveIPAddr("ip4", dest)
	c.WriteTo(buf, dst)

	recv := make([]byte, 1500)
	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		n, addr, netErr := c.ReadFrom(recv)
		if netErr != nil {
			if !strings.Contains(netErr.Error(), "timeout") {
				tuiPrint("Error: " + netErr.Error())
			}
			isMsgOutgoing = false
			setConnectedStatus(false)
			return nil
		}
		if n < 12 || recv[0] != 0 || addr.String() != dest || !bytes.Equal(recv[8:12], magic) {
			continue
		}
		setConnectedStatus(true)
		return recv[12:n]
	}
}

func listenForPackets() {
	type handler func(net.Addr, []byte) []byte
	type route struct {
		magic []byte
		fn    handler
	}

	routes := []route{
		{magicPollBytes, handlePoll},
		{magicMsgBytes, handleMessage},
		{magicHsBytes, handleHandshake},
		{magicInfoBytes, handleInfoPing},
	}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	buf := make([]byte, 1500)
	for {
		n, addr, _ := c.ReadFrom(buf)
		if n < 12 || buf[0] != 8 {
			continue
		}
		magic := buf[8:12]
		payload := make([]byte, n-12)
		copy(payload, buf[12:n])

		var matched handler
		for _, r := range routes {
			if bytes.Equal(magic, r.magic) {
				matched = r.fn
				break
			}
		}
		if matched == nil {
			reply := make([]byte, n)
			reply[0] = 0
			copy(reply[4:], buf[4:n])
			reply[2], reply[3] = 0, 0
			s := makeChecksum(reply)
			reply[2], reply[3] = byte(s>>8), byte(s)
			c.WriteTo(reply, addr)
			continue
		}
		result := matched(addr, payload)
		reply := buildPacket(magic, result)
		copy(reply[4:8], buf[4:8])
		s := makeChecksum(reply)
		reply[2], reply[3] = byte(s>>8), byte(s)
		c.WriteTo(reply, addr)
	}
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
