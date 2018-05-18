.PHONY: all
all: test

.tmp:
	mkdir -p .tmp

.PHONY: test
test: $(wildcard **/*.go) .tmp
	go test -coverprofile=.tmp/c.out ./...
	go tool cover -func=.tmp/c.out
