check:
	@go vet $(go list ./src/... | grep -v vendor)
	@go test -v ./src/...

.PHONY: check
