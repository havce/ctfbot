FROM golang:1.21.6-alpine3.19 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o ctfbotd ./cmd/ctfbotd

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/ctfbotd /ctfbotd

ENTRYPOINT ["/ctfbotd", "-config-path", "/ctfbotd.toml"]
