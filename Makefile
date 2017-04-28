.PHONY: all build check cli ci coverage dev image lint mocks release stats test vendor

VERSION ?= $(shell date +%Y%m%d%H%M%S)

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

release:
	docker build -t convox/praxis:$(VERSION) .
	docker push convox/praxis:$(VERSION)
	make -C cmd/cx release VERSION=$(VERSION)
	make -C provider release VERSION=$(VERSION)
	env VERSION=$(VERSION) bin/release

stats:
	cloc . --exclude-dir=vendor

test: check
	env FRONTEND=none govendor test +local

vendor:
	go get -u github.com/kardianos/govendor
	govendor fetch +outside
	govendor remove +unused
