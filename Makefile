build: test
	go build

test:
	go test -v ./...

verify: plugin
	./test/archive-test.sh

plugin:
	go build -buildmode=plugin -o ./test/sample.so ./plugin/sample.go

.PHONY: test build verify plugin