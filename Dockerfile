FROM golang:1.19 AS build
WORKDIR /go/src
COPY db ./db
COPY main.go .
COPY main_test.go .
COPY sql ./sql
COPY web ./web
COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0

