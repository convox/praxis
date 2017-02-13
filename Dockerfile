FROM golang:1.7.5-alpine

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./...
