.DEFAULT_GOAL := build/xpdt
GO=go

build/xpdt: $(shell find . -type f -name "*.go")
	$(GO) build -o build/xpdt -ldflags "-s -w" go.katupy.io/xpdt/cmd

go/test:
	$(GO) test -count=1 ./...
