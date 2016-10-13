.PHONY: all mocks vendor

all:

mocks:
	make -C provider mocks

vendor:
	godep save ./...
