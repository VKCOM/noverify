NOW=`date +%Y%m%d%H%M%S`
OS=`uname -n -m`
AFTER_COMMIT=`git rev-parse HEAD`

install:
	go install -ldflags "-X 'main.BuildTime=$(NOW)' -X 'main.BuildOSUname=$(OS)' -X 'main.BuildCommit=$(AFTER_COMMIT)'" .

check:
	@go vet $(go list ./src/... | grep -v vendor)
	@go test -race -v ./src/...

.PHONY: check
