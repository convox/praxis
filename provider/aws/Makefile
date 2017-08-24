.PHONY: release

release:
	mkdir -p /tmp/release/$(VERSION)
	jq '.Mappings.Constants.Values.Version|= "$(VERSION)"' formation/rack.json > /tmp/release/$(VERSION)/rack.json
	make -C lambda/lifecycle release VERSION=$(VERSION)
	aws s3 cp formation s3://praxis-releases/release/$(VERSION)/formation --recursive --acl public-read
	aws s3 cp /tmp/release/$(VERSION)/rack.json s3://praxis-releases/release/$(VERSION)/formation/rack.json --acl public-read
