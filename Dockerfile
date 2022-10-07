# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app
RUN mkdir bin

COPY go.mod ./
COPY go.sum ./
COPY .env.production ./.env

ADD internal internal
ADD cmd cmd

RUN go mod download

RUN cd cmd/server && go build -o ../../bin/server

CMD [ "./bin/server" ]