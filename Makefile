BINARY     := magik
BINARY_GUI := magik-gui
MODULE     := github.com/ehsatn/MAGIK
CMD        := ./cmd/magik
GUI_CMD    := ./cmd/magik-gui
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
BUILT_BY   := make

LDFLAGS := -s -w \
	-X $(MODULE)/pkg/version.Version=$(VERSION) \
	-X $(MODULE)/pkg/version.Commit=$(COMMIT) \
	-X $(MODULE)/pkg/version.BuildDate=$(BUILD_DATE) \
	-X $(MODULE)/pkg/version.BuiltBy=$(BUILT_BY)

GOFLAGS := -trimpath

.PHONY: all build build-gui build-all build-gui-all clean test lint fmt vet run release install

all: build

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

build-gui:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_GUI) $(GUI_CMD)

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe $(CMD)

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe $(CMD)

build-gui-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS) -H=windowsgui" -o dist/$(BINARY_GUI)-windows-amd64.exe $(GUI_CMD)

build-gui-windows-arm64:
	GOOS=windows GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS) -H=windowsgui" -o dist/$(BINARY_GUI)-windows-arm64.exe $(GUI_CMD)

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 $(CMD)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 $(CMD)

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 $(CMD)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 $(CMD)

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64
	@echo "All targets built in dist/"

build-gui-all: build-gui-windows-amd64 build-gui-windows-arm64
	@echo "GUI targets built in dist/"

install:
	go install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(CMD)

test:
	go test -race -cover ./...

test-short:
	go test -short ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY).exe $(BINARY_GUI) $(BINARY_GUI).exe
	rm -rf dist/

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean --skip=publish

deps:
	go mod tidy
	go mod verify

.DEFAULT_GOAL := build
