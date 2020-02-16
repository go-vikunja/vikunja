FROM golang:1.13.5-alpine

RUN apk update
RUN apk add alpine-sdk

WORKDIR /testfixtures
COPY . .

RUN go mod download
