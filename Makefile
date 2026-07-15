.PHONY: run test fmt build

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w api cmd config database internal migrations pkg web

build:
	mkdir -p bin
	go build -o bin/ai-argus ./cmd/server
