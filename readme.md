# PingChat v2

### **E2EE lightweight TUI chat with colored usernames, live presence, configurable identities, and ICMP-based message transport**

---
## Usage:

### **Client:**
```
sudo ./pingchat
```
It should automatically walk you through the setup process.

The `Color` field accepts [tview color names](https://pkg.go.dev/github.com/gdamore/tcell/v2#pkg-constants) or hex (`#ff8800`).

All clients sharing the same `Shared password` are in the same conversation.

Miscellaneous notes:

- Messages are end-to-end encrypted with ChaCha20-Poly1305.
- Only the most recent message per conversation is stored server-side.
- Outgoing pings are designed to be fairly discrete and difficult to detect or decode by any sniffers. 

---

### **Server:**
The server needs to be a public IPv4 address that is handling incoming ICMP pings to its IP.

The binary should be run once and made persistent (recommended with tmux or nohup)

Note that the server temporarily disables the kernel's ICMP replies while it is running, but it should automatically re-enable them upon being stopped.

Normal ICMP traffic should function normally while the server is running. The clients send specialized magic bytes to trigger custom replies, without these it just replies like a normal return ping.

Run it with:
```
sudo ./pingchat -server -salt <keyboard mashing>
```

## Flags

| Flag      | Used By | Description                                      |
|-----------|---------|--------------------------------------------------|
| `-server` | Server  | Run program as a server instance                 |
| `-salt`   | Server  | Salt that the server uses for anti-impersonation |
| `-login`  | Client  | Force re-configuration of user config info       |
| `-path`   | Client  | Path to config file to load or write to          |