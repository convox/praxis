.PHONY: all build clean release

all: lambda.zip

clean:
	rm -f lambda.zip main

lambda.zip: index.js main
	zip -r lambda.zip main index.js

main: *.go
	GOOS=linux GOARCH=amd64 go build -o main

release: lambda.zip
	aws s3 cp lambda.zip s3://praxis-releases/release/$(VERSION)/lambda/lifecycle.zip --acl public-read
	aws s3 cp lambda.zip s3://praxis-releases/release/latest/lambda/lifecycle.zip --acl public-read

test:
	bin/test $(RACK) $(INSTANCE)

upload: lambda.zip
	bin/upload $(RACK)
