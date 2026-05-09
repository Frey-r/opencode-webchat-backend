# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Run stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates opencode

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser

COPY --from=builder /app/server .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/goose .

USER appuser

EXPOSE 8080

CMD ["./server"]