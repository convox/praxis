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
	make -C provider release VERSION=$(VERSION)
	docker build -t convox/praxis:$(VERSION) .
	docker push convox/praxis:$(VERSION)
	@echo "release: $(VERSION)"

stats:
	cloc . --exclude-dir=vendor

test: check
	env FRONTEND=none govendor test +local

vendor:
	go get -u github.com/kardianos/govendor
	govendor fetch +outside
	govendor remove +unused
