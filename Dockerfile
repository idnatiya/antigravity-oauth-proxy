# syntax=docker/dockerfile:1

# Stage 1: Build binary
FROM golang:alpine AS builder

WORKDIR /build

# Install git and ca-certificates needed for downloading dependencies
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build standalone proxy and auth helper binaries
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/antigravity-oauth-proxy ./cmd/antigravity-oauth-proxy && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/auth ./cmd/auth

# Stage 2: Minimal runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/antigravity-oauth-proxy /app/antigravity-oauth-proxy
COPY --from=builder /build/auth /app/auth

# Default port
ENV PORT=9878
EXPOSE 9878

ENTRYPOINT ["/app/antigravity-oauth-proxy"]
