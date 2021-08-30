build:
	cd ./wasm && $(MAKE) build
	tsc -p ./www/ts/tsconfig.json

build_watch:
	cd ./wasm && $(MAKE) build
	tsc -w -p ./www/ts/tsconfig.json

run: build
	cd ./server && go run server.go
