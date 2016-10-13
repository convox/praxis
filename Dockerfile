FROM convox/golang

# define app root
ENV APP github.com/convox/praxis

# copy app source
WORKDIR $GOPATH/src/$APP
COPY . $GOPATH/src/$APP

# compile app
RUN go install ./api
