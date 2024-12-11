include .env

target := dist/index.js
src.ts := $(wildcard src/*.ts src/*/*.ts src/*/*/*.ts)
plugins := src/plugins/binary.wasm $(wildcard src/plugins/*.js)

PROTO_DST := ./src/pb
PROTO_SRC := ../../server/pb
NODE_PROTOC := npx grpc_tools_node_protoc
PROTOC_GEN_TS := node_modules/.bin/protoc-gen-ts

proto := $(wildcard $(PROTO_SRC)/*.proto)
pb.js := $(proto:$(PROTO_SRC)%.proto=$(PROTO_DST)%_pb.js)
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

$(target): prisma/schema.prisma $(pb.js) $(src.ts) .env $(plugins)
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

$(pb.js): $(proto)
	mkdir -p ./src/pb
	$(NODE_PROTOC) \
	  --plugin=protoc-gen-ts="$(PROTOC_GEN_TS)" \
	  --js_out="import_style=commonjs,binary:$(PROTO_DST)" \
	  --grpc_out="grpc_js:$(PROTO_DST)" \
	  --ts_out="service=grpc-node,mode=grpc-js:$(PROTO_DST)" \
	  -I $(PROTO_SRC) $(proto)
