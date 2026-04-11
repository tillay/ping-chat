# PingChat v2

ICMP-based encrypted terminal chat. Requires root.

## Build

```
go build -o pingchat .
```

## Usage

**Server** (run once, on the machine clients connect to):
```
sudo ./pingchat -server
```

**Client:**
```
sudo ./pingchat -ip <server_ip> -pass <shared_password> -user <username> -color <color>
```

`-color` accepts [tview color names](https://pkg.go.dev/github.com/gdamore/tcell/v2#pkg-constants) or hex (`#ff8800`).

All clients sharing the same `-pass` are in the same conversation.

## Flags

| Flag      | Default   | Description     |
|-----------|-----------|-----------------|
| `-server` | false     | Run as server   |
| `-ip`     | 127.0.0.1 | Server IP       |
| `-pass`   | _(empty)_ | Shared password |
| `-user`   | guest     | Display name    |
| `-color`  | #ffffff   | Username color  |

## Notes

- The server suppresses kernel ICMP replies on start. Run `-reauthKernel` to restore them after. (It also automatically resets on reboot)
- Messages are end-to-end encrypted with ChaCha20-Poly1305.
- Only the most recent message per conversation is stored server-side.