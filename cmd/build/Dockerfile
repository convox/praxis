FROM golang:1.7-alpine

RUN apk update && apk add docker openssh

COPY . $GOPATH/src/github.com/convox/praxis
RUN go install github.com/convox/praxis/cmd/build
