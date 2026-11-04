.PHONY: build test vet

build:
	go build -o bin/toolguard ./cmd/toolguard

test:
	go test ./...

vet:
	go vet ./...
