.PHONY: all build check ci coverage dev lint mocks test test-deps vendor

all: build

build:
	docker build -t convox/praxis .

check:
	bin/check

ci: test
	bin/ci

coverage: ci
	open https://codecov.io/github/ddollar/praxis/commit/$$(git rev-parse HEAD)

dev: build
	go install ./cmd/cx
	cx rack start

lint:
	bin/lint

mocks:
	make -C provider mocks

test: test-deps check
	bin/test

test-deps:
	go get -t ./...

vendor:
	go get -u github.com/tools/godep
	go get -d -t -u ./...
	godep save ./...
