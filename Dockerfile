## convox:development

FROM convox/golang

WORKDIR $GOPATH/src/github.com/convox/praxis

COPY . .

CMD ["rerun", "-watch", ".", "-build", "github.com/convox/praxis/cmd/rack"]

## convox:production

WORKDIR $GOPATH/src/github.com/convox/praxis

RUN go install ./cmd/...

CMD ["rack"]
