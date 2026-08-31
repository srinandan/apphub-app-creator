# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL := /bin/bash
.DEFAULT_GOAL := help

# Configuration variables
BINARY_NAME     ?= apphub-app-creator
PORT            ?= 8080
FRONTEND_DIR    := frontend
WEBDIST_DIR     := internal/cmd/webdist
DOCKER_IMAGE    ?= ghcr.io/srinandan/apphub-app-creator:latest
GO              ?= go
NPM             ?= npm

# ------------------------------------------------------------------------------
# Running Components
# ------------------------------------------------------------------------------

.PHONY: start dev
start: ## Start both backend server and frontend development server concurrently
	@echo "Starting App Hub Creator backend (port $(PORT)) and frontend dev server..."
	@trap 'kill 0' INT TERM EXIT; \
		($(GO) run ./cmd/apphub-app-creator/apphub-app-creator.go server --port=$(PORT)) & \
		($(NPM) --prefix $(FRONTEND_DIR) run dev) & \
		wait

dev: start

.PHONY: start-backend server backend
start-backend: ## Start the Go HTTP backend server (default port 8080, override with PORT=<port>)
	@echo "Starting backend server on port $(PORT)..."
	$(GO) run ./cmd/apphub-app-creator/apphub-app-creator.go server --port=$(PORT)

server: start-backend
backend: start-backend

.PHONY: start-frontend ui frontend
start-frontend: ## Start the frontend Vite development server
	@echo "Starting frontend Vite development server..."
	$(NPM) --prefix $(FRONTEND_DIR) run dev

ui: start-frontend
frontend: start-frontend

# ------------------------------------------------------------------------------
# Building
# ------------------------------------------------------------------------------

.PHONY: build
build: ## Build the standalone Go CLI and backend binary
	@echo "Building binary $(BINARY_NAME)..."
	$(GO) build -o $(BINARY_NAME) ./cmd/apphub-app-creator/apphub-app-creator.go

.PHONY: build-frontend build-ui
build-frontend: ## Build frontend production assets
	@echo "Building frontend production bundle..."
	$(NPM) --prefix $(FRONTEND_DIR) run build

build-ui: build-frontend

.PHONY: build-embedded build-all
build-embedded: build-frontend ## Build Go binary with embedded Web UI assets (single standalone executable)
	@echo "Staging frontend assets for binary embedding..."
	@mkdir -p $(WEBDIST_DIR)
	@rm -rf $(WEBDIST_DIR)/*
	@cp -r $(FRONTEND_DIR)/dist/* $(WEBDIST_DIR)/
	@echo "Building Go binary with embedded UI..."
	$(GO) build -tags embedui -o $(BINARY_NAME) ./cmd/apphub-app-creator/apphub-app-creator.go
	@echo "Successfully built standalone $(BINARY_NAME) with embedded UI."

build-all: build-embedded

# ------------------------------------------------------------------------------
# Dependencies & Installation
# ------------------------------------------------------------------------------

.PHONY: install install-go install-frontend
install: install-go install-frontend ## Install both Go and frontend dependencies

install-go: ## Download Go module dependencies
	@echo "Downloading Go dependencies..."
	$(GO) mod download

install-frontend: ## Install frontend npm dependencies
	@echo "Installing frontend dependencies..."
	$(NPM) --prefix $(FRONTEND_DIR) install

# ------------------------------------------------------------------------------
# Testing & Code Quality
# ------------------------------------------------------------------------------

.PHONY: test test-coverage
test: ## Run unit tests across all packages
	@echo "Running unit tests..."
	$(GO) test -v internal/cmd internal/client ./cmd/apphub-app-creator

test-coverage: ## Run unit tests and generate coverage report
	@echo "Running unit tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out internal/cmd internal/client ./cmd/apphub-app-creator
	$(GO) tool cover -func=coverage.out

# ------------------------------------------------------------------------------
# Docker
# ------------------------------------------------------------------------------

.PHONY: docker-build docker-run
docker-build: ## Build container image using Dockerfile
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run container image on port 8080
	@echo "Running Docker container on port 8080..."
	docker run -it --rm -p $(PORT):8080 $(DOCKER_IMAGE) server --port=8080

# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

.PHONY: clean
clean: ## Remove build artifacts, coverage reports, and dist directories
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) coverage.out
	@rm -rf $(FRONTEND_DIR)/dist $(WEBDIST_DIR)
	@echo "Clean complete."

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

.PHONY: help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
