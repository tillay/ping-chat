package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

const infoMagic1 byte = 0x49
const infoMagic2 byte = 0x4E

type ipInfoResponse struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func buildInfoReply(ip string) []byte {
	resp, err := http.Get("https://ipinfo.io/" + ip + "/json")
	if err != nil {
		return []byte("lookup failed")
	}
	defer resp.Body.Close()
	var info ipInfoResponse
	json.NewDecoder(resp.Body).Decode(&info)

	org := info.Org
	if idx := strings.Index(org, " "); idx != -1 {
		org = org[idx+1:]
	}

	result := info.IP + " | " + info.City + ", " + info.Region + ", " + info.Country + " | " + org
	return []byte(result)
}

func sendInfoPing(dest string) string {
	buf := make([]byte, 11)
	buf[0] = 8
	buf[9], buf[10] = infoMagic1, infoMagic2
	s := makeChecksum(buf)
	buf[2], buf[3] = byte(s>>8), byte(s)

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	processErr(err)
	defer c.Close()

	dst, _ := net.ResolveIPAddr("ip4", dest)
	c.WriteTo(buf, dst)

	recv := make([]byte, 1500)
	c.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		n, addr, err := c.ReadFrom(recv)
		if err != nil {
			return ""
		}
		if n < 11 || recv[0] != 0 || addr.String() != dest || recv[9] != infoMagic1 || recv[10] != infoMagic2 {
			continue
		}
		return string(recv[11:n])
	}
}
