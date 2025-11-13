# Dodgeball (Go)

A tiny gRPC service that simulates a dodgeball passing game. You send a list of player coordinates and a starting position/direction; it tells you how many throws happen and who ends up with the ball.

The rules are simple but picky (axis and perfect diagonal throws only), and the service is fast enough for big inputs.

---

## What’s in here

- `compute/` — the pure Go simulation engine (no networking).
- `service/` — the gRPC service that wraps the engine.
- `server/` — the `main` entry that starts the server.
- `proto_gen/` — generated Go protobufs and service stubs.
- `compute/dodgeball_test.go` — unit tests for the engine.
- `e2e/e2e_test.go` — end‑to‑end gRPC tests (behind the `e2e` build tag).

---

## How the simulation works (short version)

- Directions are the 8 compass points in clockwise order:
  - `0=N`, `1=NE`, `2=E`, `3=SE`, `4=S`, `5=SW`, `6=W`, `7=NW`.
- Start with player `startIndex` (0‑based) who just received the ball from `startDirection`.
- That player scans directions clockwise starting with the one after the incoming direction and checks all 8. For each direction, it looks for alive players exactly on that ray and throws to the nearest one.
- The thrower leaves the field; the receiver becomes current and is considered to have received the ball from the opposite direction `(dir + 4) % 8`.
- Stop when the current player can’t find anyone in any of the 8 directions.
- Result: total number of throws and the index of the last player (0‑based).

Geometry checks use exact integer comparisons:
- Axis: same `x` or same `y` with the right sign.
- Diagonal: `dx == dy` (NE/SW) or `dx == -dy` (SE/NW) with the right sign.
- Distance along a valid ray: `max(|dx|, |dy|)` (Chebyshev) to pick the nearest.

---

## Run the server

Default port is `50051`. Override with `PORT`.

Without logs:
- Windows PowerShell:
  ```powershell
  go run ./server
  ```
- Linux/macOS:
  ```bash
  go run ./server
  ```
You should see:
```
Dodgeball gRPC server running on :50051
```

With request/response logs:
- Windows PowerShell:
  ```powershell
  $env:DODGEBALL_LOG_IO = "1"
  go run ./server
  ```
- Linux/macOS:
  ```bash
  DODGEBALL_LOG_IO=1 go run ./server
  ```
Example of what logs look like when a request comes in:
```
[IO] RunSimulation request: {"players":[{"x":"0","y":"0","alive":true},...],"startDirection":7,"startIndex":4}
[IO] RunSimulation response: {"throws":4,"lastPlayer":7} (duration=145.321µs)
```

---

## The gRPC API

Service: `DodgeballService`

- `RunSimulation(SimulationInput) -> SimulationResult`

Messages (field names are lowerCamelCase in clients like Node/TS):

- `Player { x: int64, y: int64, alive: bool }`
- `SimulationInput { players: Player[], startDirection: int32, startIndex: int32 }`
- `SimulationResult { throws: int32, lastPlayer: int32 }`

Notes:
- `startIndex` is 0‑based (if your input is 1‑based, subtract 1).
- `startDirection` must be 0..7 as listed above.
- Players should start with `alive: true`.

---

## Quick client example (Go)

```go
ctx := context.Background()
conn, _ := grpc.DialContext(ctx, "127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()
client := pb.NewDodgeballServiceClient(conn)

req := &pb.SimulationInput{
    Players: []*pb.Player{
        {X: -10, Y: -10, Alive: true},
        {X: -10, Y:  10, Alive: true},
        {X:   0, Y: -10, Alive: true},
        {X:   0, Y:  10, Alive: true},
        {X:  10, Y: -10, Alive: true}, // start here (index 4)
        {X:  10, Y:  10, Alive: true},
        {X:  -9, Y: -10, Alive: true},
        {X:  -9, Y:   0, Alive: true},
    },
    StartDirection: 7, // NW
    StartIndex:     4, // 0-based
}
res, err := client.RunSimulation(ctx, req)
```

If you’re using Node with `@grpc/grpc-js` + `@grpc/proto-loader`, remember int64 fields (`x`, `y`) are strings by default unless you set loader options. The object shape should be lowerCamelCase:

```ts
const req = {
  players: [ { x: "-10", y: "-10", alive: true }, /* ... */ ],
  startDirection: 7,
  startIndex: 4,
};
client.RunSimulation(req, (err, res) => { /* ... */ });
```

---

## Testing

Unit tests (engine only):
```bash
go test -v ./compute
```

End‑to‑end tests (spins up an in‑process gRPC server):
```bash
# macOS/Linux
go test -v -tags e2e ./e2e

# with request/response logs
DODGEBALL_LOG_IO=1 go test -v -tags e2e ./e2e
```
PowerShell:
```powershell
# without logs
go test -v -tags e2e ./e2e

# with logs
$env:DODGEBALL_LOG_IO = "1"; go test -v -tags e2e ./e2e
```

---

## Troubleshooting

- I only see `players: []` and zeros in the server logs:
  - Your client likely sent an empty/default message. In JS/TS, make sure you’re using lowerCamelCase field names (`players`, `startDirection`, `startIndex`, `x`, `y`, `alive`).
  - For `@grpc/proto-loader`, `int64` fields are strings by default. Send coordinates as strings or enable `{ longs: Number }`.
  - Remember `startIndex` is 0‑based.

- Results don’t match what I expect:
  - Check the direction mapping (0..7) and that targets must be exactly axis or perfect diagonals.
  - The incoming direction is checked last; the scan always tries all 8 directions clockwise starting from the next.

---
