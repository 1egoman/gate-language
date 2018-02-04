FROM golang:1.8-alpine

WORKDIR /go/src/app
COPY server/ .

RUN go install -v

CMD ["app"]
