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
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-lobby cmd/wsnet2-lobby/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-game cmd/wsnet2-game/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-bot cmd/wsnet2-bot/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-hub cmd/wsnet2-hub/main.go

build-commit: generate
	mkdir -p bin
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-lobby-$(COMMIT) cmd/wsnet2-lobby/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-game-$(COMMIT) cmd/wsnet2-game/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-bot-$(COMMIT) cmd/wsnet2-bot/main.go
	go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)" -o bin/wsnet2-hub-$(COMMIT) cmd/wsnet2-hub/main.go

