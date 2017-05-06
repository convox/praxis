.PHONY: mocks

mocks:
	go get -u github.com/vektra/mockery/.../
	rm -f mock_Provider.go
	mockery -inpkg -name Provider
