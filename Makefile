.PHONY: all build check cli ci coverage dev image lint mocks release stats test vendor

DATE=$(shell date +%Y%m%d%H%M%S)
RELEASE ?= latest
VERSION ?= $(DATE)

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
	go get -u github.com/vektra/mockery/.../
	rm -rf mocks
	mockery -all -dir types

qa:
	# local
	sudo cx rack uninstall local
	rm -rf /Users/Shared/convox
	curl https://s3.amazonaws.com/praxis-releases/release/$(RELEASE)/cli/darwin/cx -o /usr/local/bin/cx
	chmod +x /usr/local/bin/cx
	sudo cx rack install local --version $(RELEASE)
	sleep 5
	cx version
	RACK_URL=https://localhost:5443 go test -v ./cmd/qa/...
	# aws
	cx rack install aws --name test-$(DATE) --version $(RELEASE) | tee /tmp/install-$(RELEASE).log
	export $$(tail -1 /tmp/install-$(RELEASE).log) && cx rack && go test -v ./cmd/qa/...

release:
	docker build -t convox/praxis:$(VERSION) .
	docker push convox/praxis:$(VERSION)
	docker tag convox/praxis:$(VERSION) convox/praxis:latest
	docker push convox/praxis:latest
	make -C cmd/cx release VERSION=$(VERSION)
	make -C provider release VERSION=$(VERSION)
	env VERSION=$(VERSION) bin/release

stats:
	cloc . --exclude-dir=vendor

test: check
	env govendor test +local

vendor:
	go get -u github.com/kardianos/govendor
	govendor fetch +outside
	govendor remove +unused
