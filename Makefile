build: test
	go build

test:
	go test -v ./...


.PHONY: test build