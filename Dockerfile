FROM golang:1.8-alpine

RUN apk add --update bash build-base curl docker git

RUN go get github.com/convox/rerun

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/...

CMD ["rack"]
