# net-cat (TCP Chat Server)

A simple concurrent TCP chat server written in Go (learning / reboot project) plus an experimental terminal UI prototype (under `maino/`).

## Features (current state)
- Accepts multiple TCP clients (max 10) and prompts for a unique username.
- Welcomes users with ASCII art on connect.
- Basic room + message data model scaffolding (not fully wired yet).
- Prototype terminal UI (gocui) for future interactive client (stand‑alone, not yet connected to server networking layer).

## Not Yet Implemented
- Broadcasting messages between connected clients.
- Proper client disconnect detection & cleanup.
- Room creation/join/leave mechanics exposed to users.
- Persistence / message history replay.
- Authentication / security.

## Requirements
- Go 1.24+ (module declares 1.24.2).

## Getting Started
Clone or copy the repository and run:

```bash
# Run server on default port (:8989)
go run .

# Run server on custom port
go run . 5050
```
Windows PowerShell equivalent is the same.

You should see:
```
Listening on the port :8989
```
Then connect from another terminal (Linux/macOS examples):
```
nc localhost 8989
# or
telnet localhost 8989
```
(Windows: you can use `nc` from Git Bash / WSL or install a netcat binary.)

You will be prompted:
```
Welcome to TCP-Chat!
Enter Username:
```
Enter a unique username to register.

## Project Layout
```
handlers/
  serverhandler.go  # Server loop, accept logic, connection limits
  clientHandler.go  # Username prompt & client registration
  models.go         # Data models (Client, Room, Message)
main.go             # Server entry point
maino/main.go       # Experimental TUI prototype using gocui
Notes/              # Personal learning notes
```

## Data Structures (intended design)
- Client: represents a connected user.
- Room: holds members and message history (currently unused in live flow).
- Message: timestamp + sender + content.

## Concurrency Model
- One goroutine per accepted client (currently only for initial registration flow).
- Global slices guarded by a `sync.Mutex` (see Known Issues for improvements).

## Known Issues / Technical Debt
1. Struct mismatch: `serverhandler.go` assumes `Room.Members []*Client`, `Room.History []*Message`, and a `TimeCreated` field which are NOT present in `models.go` (currently `[]Client`, `[]Message`, and no timestamp). This must be reconciled (either change models or server code) for full integration later.
2. Duplicate client state: `ClientsConnected` mirrors `len(Clients)` and can desynchronize (e.g., on early close). Prefer deriving count from the slice length under lock.
3. Capacity check pattern: increments `ClientsConnected` before verifying limit; better to check first inside a single critical section.
4. `updateClientCount` goroutine:
   - Performs blocking zero‑byte reads (will hang). Use deadlines, heartbeats, or remove and rely on explicit error handling when broadcasting.
   - Iterates & mutates `Clients` without consistently holding the mutex for the entire pass (risk of race conditions).
5. Removal of clients on normal disconnect is not yet implemented in the main handler flow.
6. Global mutable variables (`Clients`, `Rooms`) would benefit from encapsulation in a `Server` struct.
7. No message broadcast logic yet; clients only register.
8. GUI (maino) is isolated and not connected to network layer.

## Suggested Next Steps
- Decide pointer vs value semantics for `Client`, `Room`, `Message` everywhere and unify.
- Implement broadcast loop: read lines from each client and send to others.
- On read/write error, remove client immediately (no periodic probe needed).
- Replace `ClientsConnected` with `len(Clients)` under lock.
- Add graceful shutdown (context + listener close + client closes).
- Integrate TUI (maino) as an actual client connecting over TCP.
- Implement rooms: join/create commands (`/join <room>`, `/list`).

## Example Broadcast Sketch (future)
```go
for {
  line, err := reader.ReadString('\n')
  if err != nil { /* remove client & return */ }
  broadcast(fmt.Sprintf("[%s][%s]: %s", time.Now().Format(tsFmt), client.Name, line))
}
```