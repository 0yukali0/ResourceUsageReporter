PHNOY: fmt
fmt:
	go fmt .

PHONY: lint
lint:
	golangci-lint run ./...