# build stage
FROM golang:1.19-alpine AS build

ADD . /go/build
WORKDIR /go/build
RUN go build -o bitopi ./main.go

# final stage
FROM alpine:3.15

EXPOSE 80

WORKDIR /var/application
CMD [ "./bitopi" ]