.PHONY: all build check ci coverage deps lint mocks test test-deps

all: build

build:
	docker build -t convox/praxis .

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

mocks:
	make -C provider mocks

test: test-deps check
	bin/test

test-deps:
	go get -t ./...
