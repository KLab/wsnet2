include .env

target := dist/index.js
src.ts := $(wildcard src/*.ts src/*/*.ts src/*/*/*.ts)
plugins := src/plugins/binary.wasm src/plugins/binary.js src/plugins/binary.d.ts src/plugins/wasm_exec.js

PROTO_DST := ./src/pb
PROTO_SRC := ../../server/pb
NODE_PROTOC := protoc
PROTOC_GEN_TS := node_modules/.bin/protoc-gen-es

proto := $(wildcard $(PROTO_SRC)/*.proto)
pb.ts := $(proto:$(PROTO_SRC)%.proto=$(PROTO_DST)%_pb.ts)
pb.go := $(proto:%.proto=%.pb.go)

.PHONY: build re-build run re-run clean install-deps

build: install-deps $(target)

re-build: $(target)

run: install-deps re-run

re-run: $(target)
	node ./dist/index.js

clean:
	rm -rf node_modules/*
	rm -rf dist
	rm -rf prisma
	rm -rf $(PROTO_DST)

install-deps:
	npm install
	chmod +x node_modules/nexus-prisma/dist-cjs/cli/nexus-prisma.js

$(target): prisma/schema.prisma $(pb.ts) $(src.ts) .env $(plugins)
	npx tsc
	cp src/plugins/binary.wasm dist/plugins/binary.wasm

src/plugins/binary.wasm: ../../server/binary/*.go go-binary/* $(pb.go)
	cd go-binary && GOOS=js GOARCH=wasm go build -o ../$@

src/plugins/wasm_exec.js:
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" $@

%.pb.go: %.proto
	$(MAKE) -C ../../server generate

prisma/schema.prisma: $(wildcard ../../server/sql/trigger.d/db.trigger)
	if [ ! -e prisma/schema.prisma ]; then \
	  rm -rf prisma; \
	  npx prisma init --datasource-provider mysql; \
	  sed -e '/^datasource db /i generator nexusPrisma {\n  provider = "nexus-prisma"\n}\n' -i $@; \
	fi
	npm run generate

$(pb.ts): $(proto)
	mkdir -p ./src/pb
	$(NODE_PROTOC) \
	  --plugin=protoc-gen-es="$(PROTOC_GEN_TS)" \
	  --es_out="$(PROTO_DST)" \
	  --es_opt="target=ts,import_extension=js" \
	  -I $(PROTO_SRC) $(proto)
