.PHONY: all build check cli ci coverage dev image lint mocks test test-deps vendor

all: build

build:
	go install ./...

check:
	bin/check

cli:
	go install ./cmd/cx

ci: test
	bin/ci

coverage: ci
	open https://codecov.io/github/ddollar/praxis/commit/$$(git rev-parse HEAD)

dev: cli image
	cx rack start

image:
	docker build -t convox/praxis .

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
