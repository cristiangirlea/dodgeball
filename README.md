# Dodgeball Monorepo

A tiny but strict dodgeball passing simulation, implemented as a Go gRPC service with a TypeScript/Express HTTP adapter. The Go service performs the simulation; the TS adapter accepts user-friendly inputs (text or JSON) over HTTP and forwards them to the Go service via gRPC.

---

## Repository layout

- `apps/dodgeball-go/` — Go implementation
  - `compute/` — pure simulation engine (no networking)
  - `service/` — gRPC service wrapper around the engine
  - `server/` — `main` entry point to run the gRPC server
  - `proto_gen/` — generated Go protobuf/code
  - `README.md` — full details for running and using the Go service
- `apps/dodgeball-ts/` — TypeScript HTTP adapter
  - `src/server.ts` — Express server exposing `POST /simulate`
  - `src/client.ts` — TS gRPC client that calls the Go service
  - `src/parser.ts` — accepts plain text (with optional leading `T`) or JSON
  - `proto_gen/` — generated TS types/client via `ts-proto`
  - `README.md` — how to run and call the HTTP adapter
- `apps/proto/` — protobuf definitions shared by both implementations
- `tests/` — sample inputs/outputs and e2e scaffolding
  - `tests/samples/*.in` — plain-text inputs
  - `tests/samples/*.out` — expected plain-text outputs
- `Makefile` — cross-platform proto generation and useful targets (requires `protoc`)
- `LICENSE` — project license

---

## Quick start

### 1) Prerequisites

- Go 1.21+ (1.22+ recommended)
- Node.js 18+
- Protobuf compiler (`protoc`) on your PATH
- Go protobuf plugins (if you plan to regenerate Go code):
  - `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
  - `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
- TypeScript protobuf plugin (installed via npm inside the TS app):
  - `cd apps/dodgeball-ts && npm install`

You typically only need to regenerate protobufs if you change `apps/proto/*.proto`.

### 2) Generate protobufs (optional, only if you changed .proto files)

From the repo root:

```bash
# Generates Go and TS code from apps/proto/*.proto
make proto
# or just TypeScript
make ts
# or just Go
make go
```

The Makefile auto-detects Windows vs macOS/Linux and uses the proper plugin paths.

### 3) Run the Go gRPC server

Default port is `:50051`.

- Windows PowerShell:
  ```powershell
  go run ./apps/dodgeball-go/server
  ```
- macOS/Linux:
  ```bash
  go run ./apps/dodgeball-go/server
  ```

With request/response logs enabled:

- PowerShell:
  ```powershell
  $env:DODGEBALL_LOG_IO = "1"
  go run ./apps/dodgeball-go/server
  ```
- macOS/Linux:
  ```bash
  DODGEBALL_LOG_IO=1 go run ./apps/dodgeball-go/server
  ```

See `apps/dodgeball-go/README.md` for more.

### 4) Run the TypeScript HTTP adapter

In another terminal:

```bash
cd apps/dodgeball-ts
npm install
npm run dev   # runs src/server.ts on http://localhost:3000
```

Production build:

```bash
npm run build
node dist/server.js
```

Note: `npm start` in this package runs `dist/index.js` (library entry), not the HTTP server. To run the server, execute `dist/server.js` as shown above.

### 5) Try it with a sample input

With the Go server running on `127.0.0.1:50051` and the TS server on `http://localhost:3000`:

```bash
curl -s -F "input=@tests/samples/sample1.in" http://localhost:3000/simulate
```

You should get a plain-text response (one line per case), e.g.:

```
4 8
5 6
```

Large input example:

```bash
curl -s -F "input=@tests/samples/sample3.in" http://localhost:3000/simulate > out.txt
```

---

## Input formats (HTTP adapter)

The HTTP adapter accepts a multipart upload with field name `input` and auto-detects the content type:

- Plain text
  - Supports a single case or multiple cases with a leading `T` (number of test cases)
  - Each case layout:
    ```
    N
    x1 y1
    x2 y2
    ...
    xN yN
    DIRECTION STARTING_PLAYER_1BASED
    ```
  - `DIRECTION` is one of `N, NE, E, SE, S, SW, W, NW`
- JSON
  - Either a single object or an array of objects
  - Each object fields:
    - `players`: array of `[x, y]` pairs
    - `startingDirection` (also accepts `startDirection`, `direction`, `dir`): one of `N, NE, E, SE, S, SW, W, NW`
    - `startingPlayer` (also accepts `start`, `s`, `startIndex`): 1-based index

Response format: `text/plain`, one line per case: `throws lastPlayer` where `lastPlayer` is 1-based to match provided sample `.out` files.

For details, see `apps/dodgeball-ts/README.md`.

---

## gRPC API (Go service)

Service: `dodgeball.DodgeballService`

- `RunSimulation(SimulationInput) -> SimulationResult`

Messages (field names are lowerCamelCase in most clients):

- `Player { x: int64, y: int64, alive: bool }`
- `SimulationInput { players: Player[], startDirection: int32, startIndex: int32 }`
- `SimulationResult { throws: int32, lastPlayer: int32 }`

Important:
- Directions are 0..7 in clockwise order: `0=N, 1=NE, 2=E, 3=SE, 4=S, 5=SW, 6=W, 7=NW`.
- `startIndex` is 0-based in the gRPC API. The HTTP adapter accepts 1-based `startingPlayer` and converts it for you.

See `apps/dodgeball-go/README.md` for full behavior and examples.

---

## Testing

From the repo root:

- Go unit tests (engine only):
  ```bash
  go test -v ./apps/dodgeball-go/compute
  ```
- Go end-to-end tests (in-process gRPC server):
  ```bash
  # macOS/Linux
  go test -v -tags e2e ./apps/dodgeball-go/e2e

  # Windows PowerShell
  go test -v -tags e2e ./apps/dodgeball-go/e2e
  ```
- Makefile helpers:
  ```bash
  make go-test   # run all Go tests in the repo
  make e2e-go    # run gRPC e2e tests only
  make test-all  # run Go + Node e2e hooks if configured
  ```

---

## Troubleshooting

- Connection refused / 500 from HTTP adapter
  - Ensure the Go gRPC server is running on `127.0.0.1:50051`
  - Check firewall/antivirus rules for localhost ports
- Empty or unexpected results
  - Use the correct multipart field name `input` when uploading
  - For text, ensure tokens are space/newline separated and, for multi-case inputs, the first token is a positive `T`
  - For JSON, use accepted field aliases and valid direction strings
- Off-by-one confusion
  - The HTTP response prints the last player index as 1-based (to match sample outputs). The gRPC API itself uses 0-based indices.

---

## License

This repository is licensed under the terms in `LICENSE`. Component packages may include their own `LICENSE` files as well (see `apps/dodgeball-go/LICENSE` and `apps/dodgeball-ts/LICENSE`).
