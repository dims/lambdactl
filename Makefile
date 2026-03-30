VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/dims/lambdactl/cmd.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/lambdactl .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

.PHONY: build install test lint clean
