include .env

.PHONY: build re-build install-deps clean

build: install-deps re-build

re-build:
	npm run build

install-deps:
	npm install

clean:
	rm -rf dist
	rm -rf node_modules
