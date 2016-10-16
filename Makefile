.PHONY: all mocks vendor

all:

dev:
	rerun -build -watch . github.com/convox/praxis/cmd/convoxd

mocks:
	make -C provider mocks

templates:
	make -C fsync templates

vendor:
	godep save ./...
