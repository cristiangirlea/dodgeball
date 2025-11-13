# Dodgeball (TypeScript HTTP adapter)

A tiny Express server and TypeScript gRPC client for the Dodgeball simulation. It accepts an input file via HTTP, forwards the request to the Dodgeball gRPC server, and returns the results in the same plain‑text format as the sample outputs.

This package is a thin adapter around the gRPC service implemented in `apps/dodgeball-go`. It does not re‑implement the simulation; it calls the Go service over gRPC.

---

## What’s in here

- `src/server.ts` — Express server exposing `POST /simulate` (multipart file upload).
- `src/client.ts` — TypeScript gRPC client (`runSimulationTS`) used by the server.
- `src/parser.ts` — Text/JSON input parser (supports multi‑case text with optional leading `T`).
- `src/rpc.ts` — Minimal gRPC transport used by the generated client.
- `proto_gen/` — `ts-proto` generated types and client stubs from `apps/proto/dodgeball.proto`.

---

## Prerequisites

- Node.js 18+ (recommended)
- The Go gRPC server running locally at `127.0.0.1:50051`:
  - See `apps/dodgeball-go/README.md` for details.
  - Quick run from repository root:
    - Windows PowerShell:
      ```powershell
      go run ./apps/dodgeball-go/server
      ```
    - macOS/Linux:
      ```bash
      go run ./apps/dodgeball-go/server
      ```

---

## Install

From the repo root (or within `apps/dodgeball-ts`):

```bash
cd apps/dodgeball-ts
npm install
```

If you modified the `.proto` files under `apps/proto/`, regenerate TS outputs:

```bash
# from repo root
make ts
```

---

## Run the HTTP server

The HTTP server listens on `http://localhost:3000` by default.

- Development (ts-node):
  ```bash
  npm run dev
  # or
  npx ts-node src/server.ts
  ```

- Production build:
  ```bash
  npm run build
  node dist/server.js
  ```

Note: The `npm start` script in `package.json` points to `dist/index.js` (library entry). To run the HTTP server, start `dist/server.js` as shown above.

---

## HTTP API

- Endpoint: `POST /simulate`
- Content-Type: `multipart/form-data`
- Field name: `input` (the uploaded file)
- Response: `text/plain`
  - One line per test case: `throws lastPlayer` (with `lastPlayer` as 1‑based index)

### Example — text input

Text format supports either a single case or multiple cases with a leading `T` (number of cases). Each case is:

```
N
x1 y1
x2 y2
...
xN yN
DIRECTION STARTING_PLAYER_1BASED
```

- `N` — number of players
- Coordinates are integers
- `DIRECTION` — one of: `N, NE, E, SE, S, SW, W, NW`
- `STARTING_PLAYER_1BASED` — 1‑based index of the starting player

Upload with curl:

```bash
curl -s -F "input=@tests/samples/sample3.in" http://localhost:3000/simulate
```

The response is plain text matching the sample `.out` format, for example:

```
4 8
5 6
```

### Example — JSON input

You can also upload JSON instead of plain text. The server auto‑detects JSON if the content starts with `{` or `[`.

Accepted shapes:

- Single case object:
  ```json
  {
    "players": [[-10,-10], [-10,10], [0,-10], [0,10], [10,-10], [10,10], [-9,-10], [-9,0]],
    "startingDirection": "NW",
    "startingPlayer": 5
  }
  ```
- Array of cases:
  ```json
  [
    { "players": [[0,0], [10,0]], "startingDirection": "E", "startingPlayer": 1 },
    { "players": [[0,0], [0,10]], "startingDirection": "N", "startingPlayer": 1 }
  ]
  ```

Upload as a file (curl will set a filename and content type):

```bash
curl -s -F "input=@case.json" http://localhost:3000/simulate
```

Notes about JSON field names the parser accepts:
- Players must be an array of `[x, y]` pairs under `players`.
- Starting direction may be `startingDirection`, `startDirection`, `direction`, or `dir` and must be one of `N, NE, E, SE, S, SW, W, NW`.
- Starting player may be `startingPlayer`, `start`, `s`, or `startIndex` and is 1‑based.

---

## How the simulation works (summary)

The authoritative implementation runs in the Go gRPC service. Highlights:

- Directions are the 8 compass points in clockwise order:
  - `0=N`, `1=NE`, `2=E`, `3=SE`, `4=S`, `5=SW`, `6=W`, `7=NW`.
- Start with player `startIndex` (0‑based in the service; this HTTP adapter accepts 1‑based `startingPlayer` in text/JSON and converts it for you).
- The current player scans directions clockwise starting after the incoming direction and checks all 8. For each direction, it looks for alive players exactly on that ray and throws to the nearest one.
- The thrower leaves the field; the receiver becomes current and is considered to have received the ball from the opposite direction `(dir + 4) % 8`.
- Stop when the current player can’t find anyone in any of the 8 directions.
- Result: total number of throws and the index of the last player.

Geometry uses exact integer comparisons:
- Axis: same `x` or same `y` with the right sign.
- Diagonal: `dx == dy` (NE/SW) or `dx == -dy` (SE/NW) with the right sign.
- Distance along a valid ray: `max(|dx|, |dy|)` (Chebyshev) to pick the nearest.

---

## Programmatic usage

You can also use the TS client directly in Node to call the Go gRPC server.

```ts
import { runSimulationTS, autoParseInputs } from "./src"; // or from the built package entry

const [input] = autoParseInputs(`
3
0 0
10 0
0 10
E 1
`);

const result = await runSimulationTS(input); // calls 127.0.0.1:50051 by default
console.log(result); // { throws: number, lastPlayer: number }
```

Advanced: `runSimulationTS(input, address)` lets you override the gRPC target address.

---

## Configuration & logging

- HTTP server port: `3000` (hardcoded in `src/server.ts`).
- gRPC server address: defaults to `127.0.0.1:50051` in the client. You can pass a different address to `runSimulationTS` in code; the HTTP server currently uses the default.
- Set `NODE_ENV=development` to enable verbose client hex payload logs on gRPC calls.

---

## Troubleshooting

- 500 error or connection refused:
  - Ensure the Go gRPC server is running on `127.0.0.1:50051`.
  - Firewall/antivirus may block localhost ports; allow local connections.
- Empty or unexpected results:
  - Verify you uploaded the file under field name `input`.
  - For text input, ensure numbers and tokens are space/newline separated; if using multiple cases, the first token must be a positive integer `T`.
  - For JSON input, use the accepted field names and direction strings.
- Off‑by‑one confusion: the HTTP response prints the last player index as 1‑based (to match sample `.out` files).

---

## License

This package is licensed under the terms in `LICENSE`. See also the root `LICENSE` and the Go service license in `apps/dodgeball-go/LICENSE`.
