FROM golang:1.21.6-alpine3.19 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o havcebotd ./cmd/havcebotd

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/havcebotd /havcebotd

ENTRYPOINT ["/havcebotd", "-config-path", "/havcebotd.toml"]
