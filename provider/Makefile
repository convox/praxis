.PHONY: all mocks

all:

mocks:
	go get -u github.com/vektra/mockery/.../
	rm -f mock.go
	mockery -name Provider -inpkg
	mv mock_Provider.go mock.go
