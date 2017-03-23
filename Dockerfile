FROM convox/golang

RUN go get github.com/convox/rerun

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/...

CMD ["bin/rack"]
