FROM convox/golang

RUN go get github.com/convox/rerun
RUN go get github.com/kardianos/govendor

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/...

CMD ["bin/rack"]
