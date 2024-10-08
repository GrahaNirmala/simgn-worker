FROM golang:1.21.1-alpine AS builder
RUN mkdir -p /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o _output/bin/worker main.go

FROM alpine:3.18
RUN mkdir -p /app
WORKDIR /app

COPY --from=builder /app/_output/bin/worker .

RUN apk update --no-cache && \
    apk add --no-cache \
    ca-certificates \
    bash \
    curl \
    postgresql15-client

ENTRYPOINT ["./worker"]
