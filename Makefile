# ===========================================
# Proto generation for Go + TS
# ===========================================

PROTO_DIR = apps/proto
GO_OUT    = apps
TS_OUT    = apps/dodgeball-ts/proto_gen

# Get ALL proto files recursively
PROTO_SRC = $(shell find $(PROTO_DIR) -name '*.proto')

# Detect OS
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows)

# Detect protoc
PROTOC := $(shell where protoc 2>NUL || which protoc)
ifeq ($(PROTOC),)
$(error "protoc not found - install protoc and add it to PATH")
endif

GO_BIN := $(shell go env GOPATH)

# ---- Plugin paths ----
ifeq ($(findstring NT,$(UNAME_S)),NT)
	PROTOC_GEN_GO      = $(GO_BIN)\bin\protoc-gen-go.exe
	PROTOC_GEN_GO_GRPC = $(GO_BIN)\bin\protoc-gen-go-grpc.exe
	TS_PLUGIN          = apps\dodgeball-ts\node_modules\.bin\protoc-gen-ts_proto.cmd
else
	PROTOC_GEN_GO      = $(GO_BIN)/bin/protoc-gen-go
	PROTOC_GEN_GO_GRPC = $(GO_BIN)/bin/protoc-gen-go-grpc
	TS_PLUGIN          = apps/dodgeball-ts/node_modules/.bin/protoc-gen-ts_proto
endif

# Default
all: proto

# ---------- Go generation ----------
go:
	@echo "Generating Go protobuf..."
	"$(PROTOC)" \
		--proto_path=$(PROTO_DIR) \
		--plugin=protoc-gen-go="$(PROTOC_GEN_GO)" \
		--plugin=protoc-gen-go-grpc="$(PROTOC_GEN_GO_GRPC)" \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)

# ---------- TS generation ----------
ts:
	@echo "Generating TypeScript protobuf..."
	"$(PROTOC)" \
		--proto_path=$(PROTO_DIR) \
		--plugin=protoc-gen-ts_proto="$(TS_PLUGIN)" \
		--ts_proto_out=$(TS_OUT) \
		$(PROTO_SRC)

# ---------- Run all ----------
proto: go ts
	@echo "Proto generation complete."

# Clean
clean:
	rm -rf $(GO_OUT)/*
	rm -rf $(TS_OUT)/*

# Debug
print-debug:
	@echo UNAME_S = $(UNAME_S)
	@echo PROTO_SRC = $(PROTO_SRC)
	@echo GO_BIN = $(GO_BIN)
	@echo PROTOC_GEN_GO = $(PROTOC_GEN_GO)
	@echo PROTOC_GEN_GO_GRPC = $(PROTOC_GEN_GO_GRPC)
	@echo TS_PLUGIN = $(TS_PLUGIN)


# ===========================================
# Tests and E2E
# ===========================================

# Run Go unit tests (compute) and gRPC e2e tests
.PHONY: go-test
go-test:
	cd apps/dodgeball-go && go test ./...

# Run Go gRPC server locally
.PHONY: run-server
run-server:
	go run ./apps/dodgeball-go/server

# Run Node e2e tests (requires Go server; this target starts it)
.PHONY: e2e-node
e2e-node:
	node tests/e2e_node.js

# Run only Go e2e tests (in-process server)
.PHONY: e2e-go
e2e-go:
	cd apps/dodgeball-go && go test ./e2e -run TestGRPCE2E

# Run all tests
.PHONY: test-all
test-all: go-test e2e-go e2e-node
