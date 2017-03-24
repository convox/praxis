FROM convox/golang

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/...

CMD ["bin/rack"]
