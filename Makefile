-include .env
export

PORT ?= 9878
WEB_PORT ?= 19190
AIR := $(shell command -v air 2>/dev/null || echo "$(HOME)/go/bin/air")

.PHONY: all help dev dev-be dev-fe dev-all build build-fe install-fe test format clean

## help: Tampilkan daftar command yang tersedia
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## dev: Menjalankan backend (Air) & frontend (Vite) secara paralel
dev: dev-all

## dev-be: Menjalankan Go backend dengan Air live-reload
dev-be: web/dist/index.html
	@if [ ! -x "$$(command -v $(AIR))" ]; then \
		echo "Air not found. Installing github.com/air-verse/air@latest..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "Starting Go backend with Air on port $(PORT)..."
	PORT=$(PORT) $(AIR) -c .air.toml

## dev-fe: Menjalankan Vite dev server untuk Vue 3 frontend (HMR instan)
dev-fe: web/node_modules
	@echo "Starting Vite dev server at http://localhost:$(WEB_PORT)/dashboard/"
	cd web && npm run dev -- --port $(WEB_PORT)

## dev-all: Menjalankan backend (Air) & frontend (Vite) secara paralel dengan graceful shutdown
dev-all: web/node_modules web/dist/index.html
	@echo "Dev environment: UI -> http://localhost:$(WEB_PORT)/dashboard/ | API -> http://localhost:$(PORT)"
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) --no-print-directory dev-be & \
	$(MAKE) --no-print-directory dev-fe & \
	wait

## install-fe: Install dependensi npm untuk frontend web
install-fe: web/node_modules

web/node_modules: web/package.json
	cd web && npm install
	@touch $@

# Placeholder untuk go:embed saat belum build frontend
web/dist/index.html:
	@mkdir -p web/dist
	@if [ ! -f web/dist/index.html ]; then \
		echo '<!doctype html><title>Antigravity Proxy</title><p>Dev mode: jalankan make build atau buka Vite dev server.</p>' > web/dist/index.html; \
	fi

## build-fe: Compile frontend Vue 3 ke static bundle di web/dist
build-fe: web/node_modules
	cd web && npm run build

## build: Build binary Go production (termasuk embedded Vue 3 bundle)
build: build-fe
	go build -o antigravity-oauth-proxy ./cmd/antigravity-oauth-proxy

## test: Menjalankan semua unit tests Go
test: web/dist/index.html
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

## clean: Bersihkan folder temporary (tmp/), bundle frontend, dan binary build
clean:
	rm -rf tmp antigravity-oauth-proxy
