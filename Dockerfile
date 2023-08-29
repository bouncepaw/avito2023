FROM golang:1.19 AS build
WORKDIR /go/src
COPY db ./db
COPY main.go .
COPY sql ./sql
COPY web ./web

ENV CGO_ENABLED=0
RUN go get -d -v ./...
