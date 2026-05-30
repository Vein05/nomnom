SHELL := /bin/sh

WAILS ?= $(HOME)/go/bin/wails
CLI_OUT ?= bin/nomnom
DESKTOP_DIR := desktop
FRONTEND_DIR := $(DESKTOP_DIR)/frontend

.PHONY: build build-desktop desktop lint lint-root lint-desktop test test-root test-desktop-go test-frontend test-frontend-build test-desktop smoke-cli-live

build:
	@mkdir -p $(dir $(CLI_OUT))
	go build -o $(CLI_OUT) .

build-desktop:
	@test -x "$(WAILS)" || (echo "Wails CLI not found at $(WAILS). Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest" && exit 1)
	cd $(FRONTEND_DIR) && npm ci && npm run build
	cd $(DESKTOP_DIR) && $(WAILS) build -clean

desktop: build-desktop

dev:
	@test -x "$(WAILS)" || (echo "Wails CLI not found at $(WAILS). Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest" && exit 1)
	cd $(DESKTOP_DIR) && $(WAILS) dev

test-root:
	@echo "[root] go test"
	@go test ./...

test-desktop-go:
	@echo "[desktop] go test"
	@cd $(DESKTOP_DIR) && go test ./...

test-frontend:
	@echo "[frontend] npm test"
	@cd $(FRONTEND_DIR) && npm ci && npm test

test-frontend-build:
	@echo "[frontend] npm run build"
	@cd $(FRONTEND_DIR) && npm ci && npm run build

test-desktop: test-desktop-go test-frontend test-frontend-build

test: test-root test-desktop
	@echo "test passed"

smoke-cli-live:
	@echo "[root] live CLI smoke"
	@NOMNOM_LIVE_SMOKE=1 go test -run TestLiveCLISmoke -count=1 .

lint-root:
	@echo "[root] go vet"
	@go vet ./...
	@echo "[root] gofmt"
	@files=$$(find . -path ./desktop -prune -o -name '*.go' -print); \
	if [ -n "$$files" ]; then \
		out=$$(gofmt -l $$files); \
		if [ -n "$$out" ]; then \
			echo "gofmt required for:"; \
			echo "$$out"; \
			exit 1; \
		fi; \
	fi

lint-desktop:
	@echo "[desktop] go vet"
	@cd $(DESKTOP_DIR) && go vet ./...
	@echo "[desktop] gofmt"
	@cd $(DESKTOP_DIR) && files=$$(find . -path ./frontend/wailsjs -prune -o -name '*.go' -print); \
	if [ -n "$$files" ]; then \
		out=$$(gofmt -l $$files); \
		if [ -n "$$out" ]; then \
			echo "gofmt required for:"; \
			echo "$$out"; \
			exit 1; \
		fi; \
	fi

lint: lint-root lint-desktop
	@echo "lint passed"
