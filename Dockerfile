FROM golang:1.19.0-alpine3.16 AS build

WORKDIR /app

COPY db ./
COPY go ./
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY main_test.go ./

EXPOSE 8080
