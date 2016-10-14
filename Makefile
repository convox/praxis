.PHONY: all mocks vendor

all:

mocks:
	make -C provider mocks

templates:
	make -C fsync templates

vendor:
	godep save ./...
