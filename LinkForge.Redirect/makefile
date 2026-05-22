.PHONY: build run test lint tidy clean

BINARY=bin/redirect
MAIN=./cmd/redirect

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

test: 
	go test -race -cover ./...

lint:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out