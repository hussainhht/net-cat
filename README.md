# net-cat (TCP Chat Server)

**made in Reboot Coding Institute**

A simple concurrent TCP chat server in Go with basic multi-room support and a minimal admin TUI (terminal UI) for monitoring.

## Features
- Concurrent TCP server with up to 10 clients (thread-safe admission).
- Welcome banner and username prompt (defaults to Anonymous; adds suffix for duplicates).
- Rooms: prompt on connect to join or create; per-room member list.
- Broadcasting: messages go to everyone in the same room.
- History sync: new joiners receive prior room messages.
- Join/leave announcements (SERVER messages).
- Commands: `/help`, `/who` (alias `/list`), `/exit`.
- Admin TUI (`handlers/guiHandler.go`): view rooms and members, navigate, and kick users.

## Requirements
- Go 1.24+ (module declares 1.24.2).

## Run
```bash
# Default port (:8989)
go run .

# Custom port
go run . 5050
```
On Windows PowerShell, the commands are the same. The admin TUI starts automatically; press Ctrl+C to exit the UI (the program will exit).

You should see:
```
Listening on the port :8989
```

## Connect
Use any TCP client:
```bash
# macOS/Linux
nc localhost 8989
# or
telnet localhost 8989
```
On Windows, install a netcat binary (or use Git Bash/WSL) or use `telnet` if available.

You’ll be prompted for a username and a room name.

## Commands (type into the chat)
- `/help` – show help
- `/who` or `/list` – list members in your current room
- `/exit` – disconnect

## Admin TUI (experimental)
The admin panel appears on start:
- Rooms (left), Users (middle), Log (right)
- Arrow keys to switch rooms/users
- Ctrl+D to kick the selected user
- Ctrl+C to quit the UI

Note: The UI is an admin view running in-process; closing it exits the server.

## Project Layout
```
handlers/
  serverhandler.go   # Listener, accept loop, room creation
  clientHandler.go   # Prompts, registration, read loop, broadcast/cleanup
  models.go          # Client, Room, Message types + helpers
  HandleCommand.go   # Slash-command parsing and handlers
  guiHandler.go      # Admin TUI (gocui)
main.go              # Entry point (starts server + TUI)
Notes/               # Personal notes
```

## Architecture & Concurrency
- One goroutine per client connection (read/broadcast loop).
- Global registries:
  - `Clients []*Client` protected by `ClientsMutex`.
  - `Rooms []*Room` protected by `RoomsMutex`.
- Admission control: capacity check uses `len(Clients)` under `ClientsMutex`.
- Broadcast: message stored under `RoomsMutex`, then delivered to members outside the lock.
- Cleanup: on read/write errors and `/exit`, client is removed from room and global list.

## Configuration
- Default port is `:8989`. Pass a port as the first CLI arg to override, e.g. `go run . 5050`.
- Max clients is `10` (constant `maxClients` in `serverhandler.go`).

## Development
```bash
go build ./...
go vet ./...
```

## Troubleshooting
- “Server is full…”: reached `maxClients` concurrent connections.
- Can’t connect with `nc` on Windows: install netcat (e.g., via Git for Windows) or use WSL.
- UI not updating: it refreshes every second; ensure terminal window is large enough.

## License
Educational/personal project.