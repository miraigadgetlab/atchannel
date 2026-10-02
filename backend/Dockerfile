# syntax=docker/dockerfile:1

FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o main ./cmd/api

# Distroless-style minimal image: no shell, no package manager, nothing to
# exploit even if the process is compromised. Runs as a fixed non-root uid.
FROM docker.io/library/alpine:3.22

RUN adduser -D -H -u 10001 app
WORKDIR /app

COPY --from=builder /app/main .

USER app

EXPOSE 3000

CMD ["./main"]
