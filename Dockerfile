# syntax=docker/dockerfile:1

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:alpine AS builder
WORKDIR /build

# Install git and ca-certificates needed for downloading dependencies
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
# Copy built frontend into web/dist for go:embed
COPY --from=frontend-builder /app/web/dist ./web/dist

# Build standalone proxy binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/antigravity-oauth-proxy ./cmd/antigravity-oauth-proxy

# Stage 3: Minimal runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/antigravity-oauth-proxy /app/antigravity-oauth-proxy

# Default port
ENV PORT=9878
EXPOSE 9878

ENTRYPOINT ["/app/antigravity-oauth-proxy"]
