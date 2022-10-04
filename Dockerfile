# build stage
FROM golang:1.19 AS build

ADD . /go/build
WORKDIR /go/build
RUN go build -o bitopi ./main.go

# final stage
FROM alpine:3.15

COPY --from=build /go/build/config /var/application/config

EXPOSE 8001

WORKDIR /var/application
CMD [ "./bitopi" ]