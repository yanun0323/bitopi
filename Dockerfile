# build stage
FROM golang:1.19-alpine AS build

ADD . /go/build
WORKDIR /go/build
# install gcc
RUN apk add build-base
# install timezone data
RUN apk add tzdata

RUN ZONEINFO=/usr/share/timezone go build -o bitopi ./main.go

# final stage
FROM alpine:3.15

COPY --from=build /go/build/bitopi /var/application/bitopi
COPY --from=build /go/build/config /var/application/config

EXPOSE 8001

WORKDIR /var/application
CMD [ "./bitopi" ]