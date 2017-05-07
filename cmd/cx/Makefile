.PHONY: release

clean:
	rm -f pkg/cx-*

release: clean
	xgo -out pkg/cx -targets 'darwin/amd64,linux/amd64' -ldflags "-X main.Version=$(VERSION)" github.com/convox/praxis/cmd/cx
	upx -q pkg/cx-darwin-10.6-amd64
	upx -q pkg/cx-linux-amd64
	aws s3 cp pkg/cx-darwin-10.6-amd64 s3://praxis-releases/release/$(VERSION)/cli/darwin/cx --acl public-read
	aws s3 cp s3://praxis-releases/release/$(VERSION)/cli/darwin/cx s3://praxis-releases/cli/darwin/cx --acl public-read
	aws s3 cp pkg/cx-linux-amd64 s3://praxis-releases/release/$(VERSION)/cli/linux/cx --acl public-read
	aws s3 cp s3://praxis-releases/release/$(VERSION)/cli/linux/cx s3://praxis-releases/cli/linux/cx --acl public-read
