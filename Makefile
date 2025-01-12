BINARY=rate-limiter

test:
	go test -v -cover -race -covermode=atomic ./...

build:
	go build -o bin/${BINARY} main.go

unittest:
	go test -short  ./...

clean:
	if [ -f bin/${BINARY} ] ; then rm -rf bin/${BINARY} ; fi

lint-prepare:
	@echo "Installing golangci-lint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.63.4
	golangci-lint --version

lint:
	golangci-lint run ./...

.PHONY: unittest build lint-prepare lint test
