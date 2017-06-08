## development ###############################################################

FROM convox/golang AS development

ENV DEVELOPMENT=true

WORKDIR $GOPATH/src/github.com/convox/praxis

COPY . .

CMD ["rerun", "-watch", ".", "-build", "github.com/convox/praxis/cmd/rack"]

## production ################################################################

FROM convox/golang AS production

ENV DEVELOPMENT=false

WORKDIR $GOPATH/src/github.com/convox/praxis

COPY --from=development $GOPATH/src/github.com/convox/praxis $GOPATH/src/github.com/convox/praxis
RUN go install ./cmd/...

CMD ["rack"]
