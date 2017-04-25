.PHONY: release

release:
	xgo -out pkg/cx -targets 'darwin/amd64,linux/amd64' github.com/convox/praxis/cmd/cx
	aws s3 cp pkg/cx-darwin-10.6-amd64 s3://praxis-releases/release/$(VERSION)/cli/macos/cx --acl public-read
	aws s3 cp s3://praxis-releases/release/$(VERSION)/cli/macos/cx s3://praxis-releases/cli/macos/cx --acl public-read
	aws s3 cp pkg/cx-linux-amd64 s3://praxis-releases/release/$(VERSION)/cli/linux/cx --acl public-read
	aws s3 cp s3://praxis-releases/release/$(VERSION)/cli/linux/cx s3://praxis-releases/cli/linux/cx --acl public-read
