package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func magicPool(sauce string, hour int64) []byte {
	seed := sha256.Sum256([]byte(sauce + strconv.FormatInt(hour, 10)))
	pool := seed[:]
	for len(pool) < 512 {
		next := sha256.Sum256(pool)
		pool = append(pool, next[:]...)
	}
	return pool[:512]
}

func currentHour() int64 {
	return time.Now().UTC().Unix() / 3600
}

func randomMagic(sauce string) []byte {
	pool := magicPool(sauce, currentHour())
	offset := rand.Intn(len(pool)/4) * 4
	return pool[offset : offset+4]
}

func matchSauce(magic []byte, sauces []string) string {
	hour := currentHour()
	for _, h := range []int64{hour, hour - 1} {
		for _, s := range sauces {
			pool := magicPool(s, h)
			for i := 0; i+4 <= len(pool); i += 4 {
				if bytes.Equal(magic, pool[i:i+4]) {
					return s
				}
			}
		}
	}
	return ""
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
}

var icmpSeq uint32

func buildPacket(magic []byte, data []byte) []byte {
	buf := make([]byte, 12+len(data))
	buf[0] = 8
	seq := uint16(atomic.AddUint32(&icmpSeq, 1))
	buf[6], buf[7] = byte(seq>>8), byte(seq)
	copy(buf[8:12], magic)
	copy(buf[12:], data)
	s := makeChecksum(buf)
	buf[2], buf[3] = byte(s>>8), byte(s)
	return buf
}

func sendPacket(sauce string, data []byte, dest string) []byte {
	magic := randomMagic(sauce)
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
		if n < 12 || recv[0] != 0 || addr.String() != dest || matchSauce(recv[8:12], []string{sauce}) == "" {
			continue
		}
		setConnectedStatus(true)
		return recv[12:n]
	}
}

func listenForPackets() {
	type handler func(net.Addr, []byte) []byte

	handlers := map[string]handler{
		"poll":  handlePoll,
		"msg":   handleMessage,
		"shake": handleHandshake,
		"info":  handleInfoPing,
	}
	sauces := []string{"poll", "msg", "shake", "info"}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	// do not touch. it will break. this code is hanging on by a thread
	buf := make([]byte, 1500)
	for {
		n, addr, _ := c.ReadFrom(buf)
		if n < 12 || buf[0] != 8 {
			continue
		}
		magic := make([]byte, 4)
		copy(magic, buf[8:12])
		payload := make([]byte, n-12)
		copy(payload, buf[12:n])

		sauce := matchSauce(magic, sauces)
		if sauce == "" {
			reply := make([]byte, n)
			reply[0] = 0
			copy(reply[4:], buf[4:n])
			reply[2], reply[3] = 0, 0
			s := makeChecksum(reply)
			reply[2], reply[3] = byte(s>>8), byte(s)
			c.WriteTo(reply, addr)
			continue
		}
		result := handlers[sauce](addr, payload)
		reply := buildPacket(magic, result)
		fmt.Printf("Replying %s (%d bytes)\n", result, len(reply))
		reply[0] = 0
		copy(reply[4:8], buf[4:8])
		reply[2], reply[3] = 0, 0
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
