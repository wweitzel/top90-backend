# syntax=docker/dockerfile:1

FROM golang:1.23-alpine

WORKDIR /app

RUN mkdir bin

COPY go.mod .
COPY go.sum .
COPY internal internal
COPY cmd cmd

RUN go mod download
RUN cd cmd/api && go build -o ../../bin/api

CMD [ "./bin/api" ]