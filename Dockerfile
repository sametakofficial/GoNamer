FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o gonamer ./cmd/gonamer

RUN chown -R 0:0 /app

FROM debian:bullseye-slim AS runner

WORKDIR /app/cache

RUN apt-get update && apt-get install -y \
    bash \
    coreutils \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/gonamer /usr/local/bin/gonamer

COPY --from=builder /app/config.yml /app/config.yml

VOLUME ["/media", "/app/cache"]

USER 0:0

ENTRYPOINT ["/usr/local/bin/gonamer"]
CMD ["rename", "/media", "--config", "/app/config.yml", "--dry-run"]
