build: test
	go build

test:
	go test -v ./...

verify:
	./test/archive-test.sh

plugin:
    go build
.PHONY: test build verify