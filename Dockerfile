# FROM golang:1.8-alpine
FROM golang:1.8

# RUN apk add --update git curl build-base
RUN apt-get update
RUN apt-get install -y git curl build-essential

WORKDIR /go/src/app
COPY server/ .

RUN go get -u github.com/radovskyb/watcher/...
RUN go get -u github.com/gorilla/websocket

RUN go install -v

CMD ["app"]
