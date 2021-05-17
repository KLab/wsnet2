# binaries to build
TARGETS := bin/wsnet2-lobby bin/wsnet2-game bin/wsnet2-hub bin/wsnet2-bot

# dependencies
PKG_LOBBY := cmd/wsnet2-lobby lobby lobby/service    auth binary common config log pb
PKG_GAME  := cmd/wsnet2-game  game  game/service     auth binary common config log pb
PKG_HUB   := cmd/wsnet2-hub   hub   hub/service game auth binary common config log pb
PKG_BOT   := cmd/wsnet2-bot   lobby lobby/service    auth binary common config log pb

# protoc targets
proto := $(wildcard pb/*.proto)
pb.go := $(proto:.proto=.pb.go)

# stringer targets
stringer := $(shell grep -r '^//go:generate stringer' | \
                sed -E 's/^([^:]*):.* -type=([0-9a-zA-Z]*)( .*)?$$/\1>\L\2_string.go/g')
string.go := $(foreach s,$(stringer),$(dir $(word 1,$(subst >, ,$(s))))$(word 2,$(subst >, ,$(s))))

COMMIT := $(shell git rev-parse --short HEAD)
GOBUILD := go build -ldflags "-X main.WSNet2Version=v0.0.1 -X main.WSNet2Commit=$(COMMIT)"

.PHONY: all generate clean test check install-go-tools build build-commit

all: install-go-tools generate build

generate: $(pb.go) $(string.go)

clean:
	$(RM) pb/*.pb.go
	$(RM) **/*_string.go

test: generate
	go vet ./...
	go test ./...

check: install-go-tools
	staticcheck ./...

install-go-tools:
	cat tools.go | awk -F'"' '/_/ {print $$2}' | xargs -tI {} go install {}

build: $(TARGETS)

build-commit: $(TARGETS:%=%-$(COMMIT))

%-$(COMMIT): %
	cp $< $@

bin/wsnet2-lobby: $(PKG_LOBBY:%=%/*.go) $(pb.go) $(string.go)
	mkdir -p bin
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-game: $(PKG_GAME:%=%/*.go) $(pb.go) $(string.go)
	mkdir -p bin
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-hub: $(PKG_HUB:%=%/*.go) $(pb.go) $(string.go)
	mkdir -p bin
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

bin/wsnet2-bot: $(PKG_BOT:%=%/*.go) $(pb.go) $(string.go)
	mkdir -p bin
	$(GOBUILD) -o $@ $(@:bin/%=./cmd/%)

%.pb.go: %.proto
	protoc --proto_path=pb --go_out=plugins=grpc:pb "$<"
	protoc-go-inject-tag --input="$@"

define stringer_rule
$(eval _1 := $(word 1,$(1)))
$(eval _2 := $(word 2,$(1)))
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
