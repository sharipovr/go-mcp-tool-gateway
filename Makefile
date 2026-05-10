.PHONY: run build test clean docker

run:
	go run ./cmd/server

build:
	go build -o bin/mcp-gateway ./cmd/server

test:
	go test ./... -v -count=1

clean:
	rm -rf bin/

docker:
	docker build -t mcp-tool-gateway:step-01 .
