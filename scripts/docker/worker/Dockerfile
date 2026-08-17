## build
FROM golang:1.26.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY cmd/main.go .
RUN go build -o /ledger-worker

## deploy
FROM ubuntu AS base

WORKDIR /

COPY --from=build /ledger-worker /ledger-worker

EXPOSE 5002

ENTRYPOINT [ "/ledger-worker" ]