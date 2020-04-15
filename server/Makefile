proto := $(wildcard pb/*.proto)
pb.go := $(proto:.proto=.pb.go)

all: $(pb.go)

clean:
	$(RM) $(pb.go)

test: $(pb.go)
	go test ./...

%.pb.go: %.proto
	protoc --proto_path=pb --go_out=plugins=grpc:pb "$^"
	protoc-go-inject-tag --input="$@"
