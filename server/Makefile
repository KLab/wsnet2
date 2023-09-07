# binaries to build
TARGETS := bin/wsnet2-lobby bin/wsnet2-game bin/wsnet2-hub bin/wsnet2-bot bin/wsnet2-tool
VERSION := $(shell git describe --tag 2>/dev/null || echo "v0.0.0")

# dependencies
PKG_LOBBY := . cmd/wsnet2-lobby lobby lobby/service auth binary common config log pb
PKG_GAME  := . cmd/wsnet2-game  game  game/service  auth binary common config log pb
PKG_HUB   := . cmd/wsnet2-hub   hub   hub/service   auth binary common config log pb game client
PKG_BOT   := . cmd/wsnet2-bot   lobby lobby/service auth binary common config log pb
PKG_TOOL  := . cmd/wsnet2-tool cmd/wsnet2-tool/cmd       binary        config     pb

# protoc targets
proto := $(wildcard pb/*.proto)
pb.go := $(proto:.proto=.pb.go)

# stringer targets
stringer := $(shell grep -r '^//go:generate stringer' --include='*.go' . | \
                sed -E 's/^\.\/([^:]*):.* -type=([0-9a-zA-Z]*).*$$/\1>\2_string.go/g')
string.go := $(foreach s,$(stringer),\
                 $(dir $(word 1,$(subst >, ,$(s))))$(shell echo $(word 2,$(subst >, ,$(s))) | tr A-Z a-z))

COMMIT := $(shell git rev-parse --short HEAD)
GOBUILD := go build -ldflags "-X wsnet2.Version=$(VERSION)"

export GOBIN := $(abspath bin)
export PATH := $(GOBIN):$(PATH)

.PHONY: all generate clean test check install-deps build build-commit

all: install-deps build

generate: install-deps $(pb.go) $(string.go)

clean:
	$(RM) pb/*.pb.go
	$(RM) **/*_string.go
	$(RM) bin/*

test: generate
	go vet ./...
	staticcheck ./...
	go test ./...

install-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/favadi/protoc-go-inject-tag@v1.4.0
	go install golang.org/x/tools/cmd/stringer@v0.12.0
	go install honnef.co/go/tools/cmd/staticcheck@v0.4.5

build: $(TARGETS)

build-version: $(TARGETS:%=%-$(VERSION))

%-$(VERSION): %
	cp $< $@

bin/wsnet2-lobby: $(PKG_LOBBY:%=%/*.go) $(pb.go) $(string.go)
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-game: $(PKG_GAME:%=%/*.go) $(pb.go) $(string.go)
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-hub: $(PKG_HUB:%=%/*.go) $(pb.go) $(string.go)
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-bot: $(PKG_BOT:%=%/*.go) $(pb.go) $(string.go)
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-tool: $(PKG_TOOL:%=%/*.go) $(pb.go) $(string.go)
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

%.pb.go: %.proto
	protoc --proto_path=pb --go_out=module=wsnet2:. --go-grpc_out=module=wsnet2:. "$<"
	protoc-go-inject-tag --input="$@"

define stringer_rule
$(eval _1 := $(word 1,$(1)))
$(eval _2 := $(shell echo $(word 2,$(1)) | tr A-Z a-z))
$(dir $(_1))$(_2): $(_1)
	go generate ./$(dir $(_1)).
endef

$(foreach s,$(stringer),$(eval $(call stringer_rule,$(subst >, ,$(s)))))

# quick build
.PHONY: lobby game hub bot
lobby: bin/wsnet2-lobby
game: bin/wsnet2-game
hub: bin/wsnet2-hub
bot: bin/wsnet2-bot
tool: bin/wsnet2-tool
