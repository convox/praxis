## convox:development

FROM convox/golang

WORKDIR $GOPATH/src/github.com/convox/praxis

COPY . .

RUN go install ./cmd/...

CMD ["rerun", "-watch", ".", "-build", "github.com/convox/praxis/cmd/rack"]

## convox:production

WORKDIR $GOPATH/src/github.com/convox/praxis

CMD ["rack"]
