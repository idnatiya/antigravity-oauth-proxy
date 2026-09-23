-include .env
export

PORT ?= 9878
AIR := $(shell command -v air 2>/dev/null || echo "$(HOME)/go/bin/air")

.PHONY: all help dev dev-be dev-fe dev-all build test format clean

## help: Tampilkan daftar command yang tersedia
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## dev: Menjalankan backend dengan Air (live-reload) & dashboard
dev: dev-be

## dev-be: Menjalankan Go backend dengan Air live-reload
dev-be:
	@if [ ! -x "$$(command -v $(AIR))" ]; then \
		echo "Air not found. Installing github.com/air-verse/air@latest..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "Starting Go backend with Air on port $(PORT)..."
	PORT=$(PORT) $(AIR) -c .air.toml

## dev-fe: Buka dashboard UI di browser (/dashboard)
dev-fe:
	@echo "Opening dashboard at http://localhost:$(PORT)/dashboard"
	@open "http://localhost:$(PORT)/dashboard" 2>/dev/null || xdg-open "http://localhost:$(PORT)/dashboard" 2>/dev/null || true

## dev-all: Menjalankan BE & membuka FE secara otomatis dengan graceful shutdown
dev-all:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-be & \
	sleep 1 && $(MAKE) --no-print-directory dev-fe & \
	wait

## build: Build binary proxy untuk production
build:
	go build -o antigravity-oauth-proxy ./cmd/antigravity-oauth-proxy

## test: Menjalankan semua unit tests
test:
	go test -v ./...

## format: Format kode Go menggunakan gofmt, goimports, dan gofumpt
format:
	@echo "Formatting Go code..."
	@go fmt ./...
	@if go tool golang.org/x/tools/cmd/goimports -version >/dev/null 2>&1 || go tool golang.org/x/tools/cmd/goimports -h >/dev/null 2>&1; then \
		go tool golang.org/x/tools/cmd/goimports -w .; \
	elif command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi
	@if go tool mvdan.cc/gofumpt -version >/dev/null 2>&1; then \
		go tool mvdan.cc/gofumpt -w .; \
	elif command -v gofumpt >/dev/null 2>&1; then \
		gofumpt -w .; \
	fi
	@echo "All code formatted!"

## clean: Bersihkan folder temporary (tmp/) dan binary hasil build
clean:
	rm -rf tmp antigravity-oauth-proxy
