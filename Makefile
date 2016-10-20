.PHONY: all mocks vendor

all:

dev:
	docker build -t praxis .
	docker run -it -p 9877:9877 -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd):/go/src/github.com/convox/praxis -v $(HOME)/.convox/praxis:/var/run/convox praxis

mocks:
	make -C provider mocks

templates:
	make -C fsync templates

vendor:
	godep save ./...
