NOW=`date '+%Y.%m.%d %H:%M:%S'`
OS=`uname -n -m`
AFTER_COMMIT=`git rev-parse HEAD`
GOPATH_DIR=`go env GOPATH`
PKG=github.com/VKCOM/noverify/src/cmd
VERSION=0.3.0

install:
	go install -ldflags "-X '$(PKG).BuildVersion=$(VERSION)' -X '$(PKG).BuildTime=$(NOW)' -X '$(PKG).BuildOSUname=$(OS)' -X '$(PKG).BuildCommit=$(AFTER_COMMIT)'" .

build-release:
	go run ./_script/release.go -build-version="$(VERSION)" -build-time="$(NOW)" -build-uname="$(OS)" -build-commit="$(AFTER_COMMIT)"

check:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_DIR)/bin v1.39.0
	@echo "running linters..."
	@$(GOPATH_DIR)/bin/golangci-lint run ./src/...
	@echo "running tests..."
	@go test -tags tracing -count 3 -race -v ./src/...
	@go test -race ./example/custom
	@echo "everything is OK"

.PHONY: check build-release
