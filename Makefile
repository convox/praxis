.PHONY: all check coverage test

all: test

check:
	@n=$$(gofmt -d -s . 2>&1); [[ -z $$n ]] || (echo $$n; exit 1)
	go vet ./...

lint:
	golint -set_exit_status ./...

coverage:
	go test -v ./... -coverprofile=/tmp/coverage.$$$$ && go tool cover -html=/tmp/coverage.$$$$

test: check
	go test -v ./...
