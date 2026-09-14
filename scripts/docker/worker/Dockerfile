## build
FROM golang:1.27.1-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ledger-worker ./cmd

## deploy
## deploy
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /ledger-worker /ledger-worker

EXPOSE 5002

ENTRYPOINT [ "/ledger-worker" ]