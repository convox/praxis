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

dev: build
	docker run -it --name=rack -p 5443:3000 -v ~/.convox/local:/var/convox -v /var/run/docker.sock:/var/run/docker.sock --rm convox/praxis

lint:
	golint -set_exit_status ./...

mocks:
	make -C provider mocks

test: test-deps check
	bin/test

test-deps:
	go get -t ./...
