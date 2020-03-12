protos := $(wildcard pb/*.proto)
pbs := $(protos:.proto=.pb.go)

all: $(pbs)

%.pb.go: %.proto
	protoc --proto_path=pb --go_out=plugins=grpc:pb "$^"
	protoc-go-inject-tag --input="$@"
