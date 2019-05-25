build: test
	go build

test:
	go test -v ./...

verify:
	./test/archive-test.sh

.PHONY: test build verify