.PHONY: build test snapshot

build:
	go build -o bin/docrot .

test:
	go test ./...

snapshot:
	goreleaser release --snapshot --clean --skip=publish
