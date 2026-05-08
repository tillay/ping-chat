package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Name      string `json:"nm"`
	Color     string `json:"cr"`
	ServerIP  string `json:"ip"`
	Password  string `json:"pw"`
	Signature string `json:"sg"`
}

func loadConfig(force bool) Config {
	if !force {
		if data, err := os.ReadFile(*path); err == nil {
			var cfg Config
			if err := json.Unmarshal(data, &cfg); err == nil {
				return cfg
			}
		}
	}

	cfg := promptConfig()
	os.MkdirAll(filepath.Dir(*path), 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(*path, data, 0600)
	return cfg
}

func saveConfig(cfg Config) error {
	os.MkdirAll(filepath.Dir(*path), 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(*path, data, 0600)
}

func promptConfig() Config {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Config will be saved to", *path)

	fmt.Print("Username: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Name color: ")
	color, _ := reader.ReadString('\n')
	color = strings.TrimSpace(color)

	fmt.Print("Server IP: ")
	ip, _ := reader.ReadString('\n')
	ip = strings.TrimSpace(ip)

	fmt.Print("Shared password: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	fmt.Print("Private signature: ")
	sig, _ := reader.ReadString('\n')
	sig = strings.TrimSpace(sig)

	return Config{
		Name:      name,
		Color:     color,
		ServerIP:  ip,
		Password:  pass,
		Signature: sig,
	}
}
