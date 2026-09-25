BINARY  := gira
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test lint fmt vet release clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

install:
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" .

test:
	go test ./...

lint: vet
	golangci-lint run ./...

fmt:
	gofmt -l -w .

vet:
	go vet ./...

release:
	./scripts/build.sh $(VERSION)

clean:
	rm -rf $(BINARY) dist
