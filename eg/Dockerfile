FROM ubuntu:16.04

RUN apt-get update && apt-get -y install curl docker.io jq git make python python-pip zip
RUN pip install awscli

RUN curl https://storage.googleapis.com/golang/go1.7.1.linux-amd64.tar.gz -O
RUN tar -C /usr/local -xzf go1.7.1.linux-amd64.tar.gz
ENV GOPATH /go
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin:/go/src/github.com/convox/release/bin

RUN go get github.com/ddollar/rerun

WORKDIR /go/src/github.com/convox/praxis/eg
COPY . /go/src/github.com/convox/praxis/eg
RUN go install .

CMD ["rerun", "-build", "github.com/convox/praxis/eg"]
