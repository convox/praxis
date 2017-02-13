.PHONY: all check ci coverage deps lint test test-deps

all: test

check:
	bin/check

ci: test
	bin/ci

coverage: ci
	open https://codecov.io/github/ddollar/praxis/commit/$$(git rev-parse HEAD)

deps:
	go get -u github.com/tools/godep
	go get -d ./...
	godep save ./...

lint:
	golint -set_exit_status ./...

test: test-deps check
	bin/test

test-deps:
	go get -t ./...
