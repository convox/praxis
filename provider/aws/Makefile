.PHONY: release

release:
	mkdir -p /tmp/release/$(VERSION)
	jq '.Parameters.Version.Default |= "$(VERSION)"' formation/rack.json > /tmp/release/$(VERSION)/rack.json
	aws s3 cp /tmp/release/$(VERSION)/rack.json s3://praxis-releases/release/$(VERSION)/formation/rack.json --acl public-read
	aws s3 cp formation/network.json s3://praxis-releases/release/$(VERSION)/formation/network.json --acl public-read
