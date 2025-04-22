NOW=`date '+%Y.%m.%d %H:%M:%S'`
OS=`uname -n -m`
AFTER_COMMIT=`git rev-parse HEAD`
GOPATH_DIR=`go env GOPATH`
BIN_NAME=noverify
PKG=github.com/VKCOM/noverify/src/cmd
VERSION=0.5.5

install:
	go install -ldflags "-X '$(PKG).BuildVersion=$(VERSION)' -X '$(PKG).BuildTime=$(NOW)' -X '$(PKG).BuildOSUname=$(OS)' -X '$(PKG).BuildCommit=$(AFTER_COMMIT)'" .

build: clear
	go build -ldflags "-X '$(PKG).BuildVersion=$(VERSION)' -X '$(PKG).BuildTime=$(NOW)' -X '$(PKG).BuildOSUname=$(OS)' -X '$(PKG).BuildCommit=$(AFTER_COMMIT)'" -o build/$(BIN_NAME)

release:
	go run ./_script/release.go -build-version="$(VERSION)" -build-time="$(NOW)" -build-uname="$(OS)" -build-commit="$(AFTER_COMMIT)"

generate_checkers_doc: build
	./build/noverify checkers-doc > docs/checkers_doc.md

playground_build:
	cd ./playground && $(MAKE) build

playground:
	cd ./playground && $(MAKE) run

check: lint test

lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_DIR)/bin v1.59.1
	@echo "running linters..."
	@$(GOPATH_DIR)/bin/golangci-lint run ./src/...
	@echo "no linter errors found"

test:
	@echo "running tests..."
	@go test -tags tracing -count 3 -race -v ./src/...
	@go test -race ./example/custom
	@echo "tests passed"

clear:
	if [ -d build ]; then rm -r build; fi

.PHONY: check release
