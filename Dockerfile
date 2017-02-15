FROM golang:1.7.5-alpine

RUN apk add --update bash build-base curl git

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/cx
RUN go install ./cmd/rack

CMD ["rack"]
