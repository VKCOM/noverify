NOW=`date +%Y%m%d%H%M%S`
OS=`uname -n -m`
AFTER_COMMIT=`git rev-parse HEAD`
GOPATH_DIR=`go env GOPATH`

install:
	go install -ldflags "-X 'main.BuildTime=$(NOW)' -X 'main.BuildOSUname=$(OS)' -X 'main.BuildCommit=$(AFTER_COMMIT)'" .

check:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/0d2da56da532f444df6827c5054f5d036bcd7096/install.sh | sh -s -- -b $(GOPATH_DIR)/bin 1.30.0
	@echo "running linters..."
	@$(GOPATH_DIR)/bin/golangci-lint run ./src/...
	@echo "running tests..."
	@go test -count 3 -race -v ./src/...
	@go test -race ./example/custom
	@echo "everything is OK"

.PHONY: check
