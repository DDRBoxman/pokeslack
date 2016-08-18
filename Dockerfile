FROM golang:1.7-alpine

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY ./main.go /go/src/app

RUN go get && go build

EXPOSE 9000

CMD ["app"]