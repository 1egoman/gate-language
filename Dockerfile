FROM golang:1.8

RUN apt-get update && apt-get install -y git curl build-essential

WORKDIR /go/src/app
COPY server/ .

RUN go get -u github.com/radovskyb/watcher/... && \
    go get -u github.com/gorilla/websocket

RUN go install -v

COPY app /app

CMD [ "sh", "-c", "/app serve --port $PORT" ]
