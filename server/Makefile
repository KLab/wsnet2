proto := $(wildcard pb/*.proto)
pb.go := $(proto:.proto=.pb.go)
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: all generate clean test install-go-tools build build-commit

all: build

generate: install-go-tools $(pb.go)
	go generate ./...

clean:
	$(RM) pb/*.pb.go
	$(RM) **/*_string.go

test: generate
	go test ./...

%.pb.go: %.proto
	protoc --proto_path=pb --go_out=plugins=grpc:pb "$^"
	protoc-go-inject-tag --input="$@"

install-go-tools:
	cat tools.go | awk -F'"' '/_/ {print $$2}' | xargs -tI {} go install {}

build: generate
	mkdir -p bin
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-lobby cmd/wsnet-lobby/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-game cmd/wsnet-game/main.go 
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-bot cmd/wsnet-bot/main.go 

build-commit: generate
	mkdir -p bin
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-lobby-$(COMMIT) cmd/wsnet-lobby/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-game-$(COMMIT) cmd/wsnet-game/main.go 
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet-bot-$(COMMIT) cmd/wsnet-bot/main.go 

