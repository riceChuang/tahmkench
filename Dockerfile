FROM golang:1.13 AS builder
RUN mkdir /tahmkench
WORKDIR /tahmkench
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main ./cmd

FROM alpine:3.10.3
RUN mkdir -p /tahmkench

COPY --from=builder /tahmkench/main /tahmkench/main
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /tahmkench
USER appuser

WORKDIR /tahmkench
CMD ./main
