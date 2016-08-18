FROM golang:1.7-onbuild

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY ./main.go /go/src/app

RUN go get && go build

CMD ["app"]