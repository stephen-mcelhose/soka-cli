BINARY   := soka
FIXTURE_ARBOR  := experiments/arbor-knowledge-agent/testdata/arbor-live.jsonl
FIXTURE_HARVEY := experiments/harvey-deep-research/testdata/harvey-deep-research.jsonl

.PHONY: build test vet lint check inspect experiment-arbor experiment-harvey clean

## build: compile the soka binary
build:
	go build -o $(BINARY) .

## test: run unit tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## check: build + vet + test (run before every commit)
check: build vet test

## inspect: launch the MCP Inspector UI against soka mcp
##   Opens http://localhost:5173 — use this to verify any change to cmd/mcp.go
inspect: build
	npx @modelcontextprotocol/inspector ./$(BINARY) mcp

## experiment-arbor: run the Arbor end-to-end experiment
experiment-arbor: build
	bash experiments/arbor-knowledge-agent/run.sh

## experiment-harvey: run the Harvey deep-research experiment
experiment-harvey: build
	bash experiments/harvey-deep-research/run.sh

## clean: remove built binary
clean:
	rm -f $(BINARY)
