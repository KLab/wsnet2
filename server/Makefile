proto := $(wildcard pb/*.proto)
pb.go := $(proto:.proto=.pb.go)

.PHONY: all generate clean test install-go-tools

all: generate

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
