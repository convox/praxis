.PHONY: all check ci coverage deps lint test test-deps

all: test

check:
	@n=$$(gofmt -d . 2>&1); [[ -z $$n ]] || (echo $$n; exit 1)
	go vet ./...

deps:
	go get -u github.com/tools/godep
	go get -d ./...
	godep save ./...

ci: test
	curl -s https://codecov.io/bash | env CODECOV_TOKEN=296de1c3-642d-42d7-9e4e-6ad019a0b548 bash

coverage: ci
	open https://codecov.io/github/ddollar/praxis/commit/$$(git rev-parse HEAD)

lint:
	golint -set_exit_status ./...

test: test-deps check
	bin/test

test-deps:
	go get -t ./...
