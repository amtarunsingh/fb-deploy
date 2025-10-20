APP := votes_storage
BIN := bin/$(APP)
PKG := ./...

COVERAGE_MIN := 95.0
GOLANGCI_VERSION := v2.5.0
DOCKER_DEV_FILE := docker/docker-compose.dev.yml

.PHONY: fmt fmt-check lint test test-coverage generate-mocks wire build dev-up dev-down

fmt:
	@echo ">> Running fmt"
	go fmt $(PKG)

fmt-check:
	@echo ">> Running fmt-check"
	@out=$$(gofmt -s -l .); \
	if [ -n "$$out" ]; then \
	  echo "gofmt needed in:"; echo "$$out"; exit 1; \
	fi

lint:
	@echo ">> Running linter"
	docker run --rm -v "$(PWD)":/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run ./...

test:
	@echo ">> Running tests"
	go test -race -count=1 $(PKG)

test-coverage:
	@echo ">> Running tests with coverage"
	mkdir -p tmp
	@COVERPKG=$$(go list ./... \
    		| grep -v '/internal/testlib' \
    		| grep -v '/internal/app/di' \
    		| grep -v '/internal/integration_test' \
    		| tr '\n' ',' | sed 's/,$$//'); \
	go test -race -count=1 -covermode=atomic -coverpkg="$$COVERPKG" -coverprofile=tmp/coverage.out ./...; \
	COV_LINE=$$(go tool cover -func=tmp/coverage.out | awk '/^total:/{print $$0}'); \
	COV_NUM=$$(echo $$COV_LINE | awk '{gsub(/%/,"",$$NF); print $$NF}'); \
	echo "\033[0m\033[1;34m>> To open HTML report: go tool cover -html=./tmp/coverage.out\033[0m"; \
	awk -v cov=$$COV_NUM -v min=$(COVERAGE_MIN) 'BEGIN { \
	  if (cov+0 < min) { printf "\033[31m>> Coverage too low: %.2f%% (required >= %.1f%%)\n\033[0m", cov, min; exit 1 } \
	  else { printf "\033[32m>> Coverage OK: %.2f%%\n\033[0m", cov } }'

generate-mocks:
	@echo ">> Generating mocks"
	go generate ./...

wire:
	@echo ">> Preparing DI container"
	wire ./internal/app/di

build:
	@echo ">> Build project"
	go build -o $(BIN) ./cmd/app

dev-up:
	docker compose -f $(DOCKER_DEV_FILE) up --build

dev-down:
	docker compose -f $(DOCKER_DEV_FILE) down
