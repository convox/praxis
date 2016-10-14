all: templates

templates:
	xgo -targets linux/amd64 -out watchd ./watch/cmd/watchd
	mv watchd-linux-amd64 watchd
	go-bindata -nometadata -pkg fsync -o templates.go watchd
	rm watchd
