FROM golang:1.8

# RUN apk add --update bash build-base curl docker git
RUN apt-get update && apt-get -y install build-essential curl git

# install docker
RUN apt-get install -y --no-install-recommends apt-transport-https ca-certificates software-properties-common
RUN curl -fsSL https://apt.dockerproject.org/gpg | apt-key add -
RUN add-apt-repository "deb https://apt.dockerproject.org/repo/ debian-$(lsb_release -cs) main"
RUN apt-get update && apt-get -y --no-install-recommends install docker-engine

RUN go get github.com/convox/rerun

WORKDIR $GOPATH/src/github.com/convox/praxis
COPY . .
RUN go install ./cmd/...

CMD ["bin/rack"]
