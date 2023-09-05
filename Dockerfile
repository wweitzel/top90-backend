# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app
RUN mkdir bin

COPY go.mod ./
COPY go.sum ./

ADD internal internal
ADD cmd cmd

RUN go mod download

RUN cd cmd/server && go build -o ../../bin/server

CMD [ "./bin/server" ]