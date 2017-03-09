.PHONY: all build check cli ci coverage dev image lint mocks stats test vendor

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

stats:
	cloc . --exclude-dir=vendor

test: check
	bin/test

vendor:
	go get -u github.com/kardianos/govendor
	govendor update +vendor
	govendor fetch +missing
