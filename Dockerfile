FROM golang:1.7.5-alpine

RUN apk add --update build-base curl

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./...

CMD bash
