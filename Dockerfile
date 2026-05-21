# Build stage - frontend
FROM node:alpine AS frontend-builder

WORKDIR /frontend
COPY opencode-webchat-frontend/package.json opencode-webchat-frontend/package-lock.json ./
RUN npm ci
COPY opencode-webchat-frontend/ .
RUN npm run build

# Build stage - backend
FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go mod files first for caching
COPY opencode-webchat-backend/go.mod opencode-webchat-backend/go.sum ./
RUN go mod download

COPY opencode-webchat-backend/ .

# Copy frontend build output
COPY --from=frontend-builder /frontend/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Run stage
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache ca-certificates nodejs npm curl bash unzip gcompat git

# Set environment variables for Bun and PATH
ENV BUN_INSTALL="/opt/bun"
ENV PATH="/opt/bun/bin:${PATH}"

# Install Bun directly to /opt/bun
RUN curl -fsSL https://bun.sh/install | bash

# Install opencode-ai globally
RUN bun add -g opencode-ai

# Ensure all users can read and execute Bun and its global packages
RUN chmod -R 755 /opt/bun

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