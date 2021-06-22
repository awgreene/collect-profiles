.PHONY: all build
all: build

GO := go
MODULE := $(shell $(GO) list -m)
VERSION := $(shell git describe --always --tags HEAD)$(and $(shell git status --porcelain),+$(shell scripts/worktree-hash.sh))


build:
	$(GO) build -ldflags '-X $(MODULE)/internal/version.Version=$(VERSION)' -o ./bin/collect-profiles .

install: build
	mv ./bin/collect-profiles $(shell go env GOPATH)/bin