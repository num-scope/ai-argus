.PHONY: run test fmt frontend build

ifeq ($(OS),Windows_NT)
MKDIR_BIN := if not exist bin mkdir bin
else
MKDIR_BIN := mkdir -p bin
endif

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w api cmd config database internal migrations pkg web

frontend:
	npm ci --no-audit --no-fund
	npm run build

build: frontend
	$(MKDIR_BIN)
	go build -o bin/ai-argus ./cmd/server
