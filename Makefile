.PHONY: all
all: bins test

.tmp:
	mkdir -p .tmp

.PHONY: test
test: $(wildcard **/*.go) .tmp
	go test -coverprofile=.tmp/c.out ./...
	go tool cover -func=.tmp/c.out

.PHONY: bins
bins: tools/sync/sync tools/check/check
tools/sync/sync: $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
	go build -o tools/sync/sync tools/sync/main.go
tools/check/check: $(wildcard *.go) $(wildcard */*.go) $(wildcard */*/*.go)
	go build -o tools/check/check tools/check/main.go
