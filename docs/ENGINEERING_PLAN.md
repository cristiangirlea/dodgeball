### Overview
This document outlines a pragmatic, production-lean plan to implement the “Dodgeball” problem and to verify it with unit tests and end-to-end tests via a Go gRPC backend and a Node.js client. The goal is to produce a correct and efficient solution that’s easy to reason about and maintain.

### Constraints & assumptions
- T ≤ 100 test cases; N ≤ 1000 players per test case.
- Coordinates are integers in [−1e9, 1e9].
- Directions are the eight compass points in clockwise order: N, NE, E, SE, S, SW, W, NW.
- Start direction D is the direction the starting player just received the ball from.
- Starting player index is given 1-based in the problem input; convert to 0-based internally.
- After a throw, the thrower leaves the field and the receiver becomes the next actor.
- The receiver starts scanning clockwise from the next direction after the direction they just received the ball from.

### Data structures
- Players are stored in parallel arrays (or a vector of structs):
  - xs[i], ys[i] for coordinates; alive[i] boolean.
- Integers are 64-bit in Go (`int64`) to avoid overflow when squaring deltas for distance comparisons.
- For direction logic, encode directions as integers 0..7: 0=N, 1=NE, 2=E, 3=SE, 4=S, 5=SW, 6=W, 7=NW.

### Algorithm
1. Parse input; convert the starting index to 0-based.
2. Initialize arrays for alive status and coordinates from the input.
3. Keep `cur` as the current player index and `dir` as the direction that player just received the ball from (0..7 encoding).
4. Repeat until the current player cannot throw to anyone:
   - For `step` in 1..8, compute `scanDir = (dir + step) & 7` to rotate clockwise in 45° increments starting from the next direction.
   - Among all alive players in the exact `scanDir` from `cur`, choose the nearest (minimal squared distance); if none, continue to the next direction.
   - If a target is found in some `scanDir`:
     - Increment throws.
     - Mark the current thrower as not alive.
     - Set `cur = targetIdx`.
     - Set `dir = (scanDir + 4) & 7` — the receiver has just received the ball from the opposite of the throw direction; next iteration will start scanning from the next clockwise direction.
     - Break the scanning loop and continue.
   - If no target is found in any of the 8 directions, terminate.
5. Output the total number of throws and the last player index (convert to 1-based for user-facing output; keep 0-based internally for API).

This is O(N^2) in the worst case. With N ≤ 1000, that’s fine in practice given careful implementation.

### Geometry rules
For two points A(ax, ay) and B(bx, by), compute dx = bx − ax, dy = by − ay.
- N:  dx == 0 and dy > 0 → dir = 0
- NE: dx > 0 and dy > 0 and dx == dy → dir = 1
- E:  dy == 0 and dx > 0 → dir = 2
- SE: dx > 0 and dy < 0 and dx == -dy → dir = 3
- S:  dx == 0 and dy < 0 → dir = 4
- SW: dx < 0 and dy < 0 and dx == dy → dir = 5
- W:  dy == 0 and dx < 0 → dir = 6
- NW: dx < 0 and dy > 0 and -dx == dy → dir = 7
- Otherwise, there is no exact compass-aligned direction between A and B.

Notes:
- Use 64-bit integers to avoid overflow for squared distances (dx*dx + dy*dy may reach up to ~2e18).
- Prefer squared distances for comparisons — no need for square roots.
- Collinearity is checked using simple equalities as above (no floating point rounding issues).

### Edge cases
- Multiple players in the same direction: choose the nearest by squared distance.
- No player in any of the 8 directions: the game ends.
- 45° diagonals: covered by `abs(dx) == abs(dy)` with proper sign checks.
- Large coordinates (±1e9): within int64; squared distance still fits within signed 64-bit range.
- Player removal: eliminate the thrower only; the receiver stays and acts next.
- Direction hand-off: after a throw in `scanDir`, the receiver’s `dir` becomes `(scanDir + 4) & 7`, because they just received the ball from the opposite direction.
- Starting direction: the initial `dir` provided is from where the first player got the ball; scanning begins from the next clockwise direction.

### Complexity
- Per throw: scan up to 8 directions and, for each direction, iterate all players once to pick the nearest — O(N) per throw with a small constant.
- Total throws ≤ N − 1, since each throw removes exactly one player.
- Worst-case complexity O(N^2); with N ≤ 1000 and tight implementation, this is fine.
- Memory: O(N) for coordinate and alive arrays.

### Tech stack rationale
- Go for core logic and service:
  - Strong integer performance and predictable memory usage.
  - Easy to expose via gRPC for interop with other languages.
- Node.js + TypeScript as a thin client to exercise the gRPC interface and provide an e2e demonstration.
- No caching is necessary: each throw depends on the evolving set of alive players, and building per-direction maps adds complexity without clear benefit at N ≤ 1000.
- Makefile is useful to unify common tasks (proto generation, tests, e2e). Docker is optional; with Go and Node installed locally, the repo runs deterministically enough for this scope.

### Optional Docker/Makefile
- Makefile targets (added):
  - `proto`: regenerate protobufs (Go + TS) if needed.
  - `go-test`: run Go unit tests and gRPC e2e tests.
  - `e2e-go`: run gRPC e2e tests only.
  - `run-server`: run the Go gRPC server.
  - `e2e-node`: run Node e2e against the Go server.
  - `test-all`: run everything.
- Docker could wrap the Go server and a Node client, but is not required here. If needed, a multi-stage Dockerfile for Go (builder + distroless runtime) plus a node:18-alpine image for the client would provide reproducibility.

### Implementation structure
- Go
  - `apps/dodgeball-go/compute`: pure logic (`RunSimulation`) — easy to unit test.
  - `apps/dodgeball-go/service`: gRPC service adapter using generated stubs.
  - `apps/dodgeball-go/server`: `main.go` to run the gRPC server, port configurable via `PORT` env.
  - Tests: unit tests for `compute` and in-process gRPC e2e tests.
- Node/TS
  - `apps/dodgeball-ts/proto_gen`: ts-proto generated types/service client.
  - `apps/dodgeball-ts/src/client.ts`: minimal RPC client using `@grpc/grpc-js`.
  - `tests/e2e_node.js`: starts the Go server, runs TS client against sample inputs, compares outputs.

### Notes for reviewers
- Direction encoding and the “opposite after throw” rule are the two most bug-prone parts; both are explicitly encoded and tested via the provided samples.
- All tests use integer-only math; no floats involved, avoiding precision issues.
