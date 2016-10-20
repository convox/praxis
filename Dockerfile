FROM convox/golang

RUN apt-get update && apt-get install -y docker.io

# define app root
ENV APP github.com/convox/praxis

# copy app source
WORKDIR $GOPATH/src/$APP
COPY . $GOPATH/src/$APP

# compile app
RUN go install ./api
RUN go install ./cmd/build
