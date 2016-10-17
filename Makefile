.PHONY: all mocks vendor

all:

dev:
	rerun -build github.com/convox/praxis

mocks:
	make -C provider mocks

templates:
	make -C fsync templates

vendor:
	godep save ./...
