.PHONY: help build\:dev build\:prod start\:dev stop\:dev start\:dev\:db stop\:dev\:db start\:prod\:db stop\:prod\:db attach test test\:unit test\:integration test\:e2e lint format generate clean migrate\:create migrate\:apply shell ps

.DEFAULT_GOAL := help
SHELL := /bin/bash

CYAN  = \033[0;36m
RESET = \033[0m

# ── Help ───────────────────────────────────────────────────────────────────────
help:
	@echo -e "$(CYAN)Go Starter API — available commands$(RESET)"
	@echo -e "  Local"
	@echo -e "    $(CYAN)make deps$(RESET)                Download Go dependencies"
	@echo -e "    $(CYAN)make build$(RESET)              Build production binary"
	@echo -e "    $(CYAN)make run$(RESET)                Run server locally"
	@echo -e "    $(CYAN)make dev$(RESET)                Run with air hot-reload"
	@echo -e "    $(CYAN)make lint$(RESET)               Run golangci-lint"
	@echo -e "    $(CYAN)make format$(RESET)             Run gofmt"
	@echo -e "    $(CYAN)make clean$(RESET)              Remove build artifacts"
	@echo -e "  Docker"
	@echo -e "    $(CYAN)make build:dev$(RESET)          Build development image"
	@echo -e "    $(CYAN)make build:prod$(RESET)         Build production image"
	@echo -e "    $(CYAN)make start:dev$(RESET)          Start dev stack (hot-reload, detached)"
	@echo -e "    $(CYAN)make stop:dev$(RESET)           Stop dev stack"
	@echo -e "    $(CYAN)make start:dev:db$(RESET)       Start dev database only"
	@echo -e "    $(CYAN)make stop:dev:db$(RESET)        Stop dev database"
	@echo -e "    $(CYAN)make start:prod:db$(RESET)      Start production database only"
	@echo -e "    $(CYAN)make stop:prod:db$(RESET)       Stop production database"
	@echo -e "    $(CYAN)make attach$(RESET)             Tail live app logs"
	@echo -e "    $(CYAN)make shell$(RESET)              Open shell in running dev container"
	@echo -e "  Code Generation"
	@echo -e "    $(CYAN)make generate$(RESET)           Regenerate Ent code from schema"
	@echo -e "    $(CYAN)make migrate:create$(RESET)     Generate Atlas migration from schema"
	@echo -e "    $(CYAN)make migrate:apply$(RESET)      Apply pending Atlas migrations"
	@echo -e "    $(CYAN)make ps$(RESET)                 Show running containers"
	@echo -e "  Tests"
	@echo -e "    $(CYAN)make test$(RESET)               Run all tests"
	@echo -e "    $(CYAN)make test:unit$(RESET)          Unit tests only"
	@echo -e "    $(CYAN)make test:integration$(RESET)   Integration tests only"
	@echo -e "    $(CYAN)make test:e2e$(RESET)           E2E tests only"
	@echo -e ""

# ── Local ─────────────────────────────────────────────────────────────────────
deps:
	go mod tidy
	go mod verify

build:
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server

run:
	go run ./cmd/server

dev:
	air

lint:
	golangci-lint run ./...

format:
	gofmt -w .
	goimports -w .

generate:
	go install entgo.io/ent/cmd/ent@latest
	@mkdir -p internal/shared/infrastructure/ent/generated
	@echo "package generated" > internal/shared/infrastructure/ent/generated/gen.go
	"$$(go env GOPATH)/bin/ent" generate ./internal/shared/infrastructure/ent/schema --target ./internal/shared/infrastructure/ent/generated
	@rm -f internal/shared/infrastructure/ent/generated/gen.go

clean:
	rm -rf bin/ tmp/ internal/shared/infrastructure/ent/generated/
	go clean -cache -testcache

# ── Atlas Migrations ──────────────────────────────────────────────────────────
ATLAS := "$$(go env GOPATH)/bin/atlas"

migrate\:create:
	@docker compose -f docker-compose.dev.yml exec -T db psql -U postgres -c "CREATE DATABASE atlas_dev;" 2>/dev/null || true
	$(ATLAS) migrate diff --dir "file://migrations" --to "ent://internal/shared/infrastructure/ent/schema" --dev-url "postgres://postgres:postgres@localhost:5432/atlas_dev?sslmode=disable"

migrate\:apply:
	$(ATLAS) migrate apply --dir "file://migrations" --url "postgres://postgres:postgres@localhost:5432/starter?sslmode=disable" --allow-dirty

# ── Docker ─────────────────────────────────────────────────────────────────────
build\:dev:
	docker compose -f docker-compose.dev.yml build

build\:prod:
	docker compose build

start\:dev:
	docker compose -f docker-compose.dev.yml up -d

stop\:dev:
	docker compose -f docker-compose.dev.yml down

start\:dev\:db:
	docker compose -f docker-compose.dev.yml up -d db

stop\:dev\:db:
	docker compose -f docker-compose.dev.yml stop db

start\:prod\:db:
	docker compose -f docker-compose.yml up -d db

stop\:prod\:db:
	docker compose -f docker-compose.yml stop db

attach:
	docker compose -f docker-compose.dev.yml logs -f app

shell:
	docker compose -f docker-compose.dev.yml exec app /bin/sh

ps:
	docker compose -f docker-compose.dev.yml ps

# ── Tests ──────────────────────────────────────────────────────────────────────
TEST_FILTER = grep -v '^? ' | sed -E 's/(--- PASS:.*)/\x1b[32m\1\x1b[0m/g; s/(--- FAIL:.*)/\x1b[31m\1\x1b[0m/g; s/(--- SKIP:.*)/\x1b[33m\1\x1b[0m/g; s/^(ok .*)/\x1b[32m\1\x1b[0m/g; s/^(FAIL .*)/\x1b[31m\1\x1b[0m/g'

# Run all tests (unit + integration + e2e)
test:
	@set -o pipefail; go test ./... -v -count=1 -p 1 2>&1 | $(TEST_FILTER)

# Run only unit tests (skips integration & e2e via testing.Short)
test\:unit:
	@set -o pipefail; go test ./... -v -count=1 -short 2>&1 | $(TEST_FILTER)

# Run only integration tests (filtered by test function name)
test\:integration:
	@set -o pipefail; go test ./internal/... -v -run Integration -count=1 -p 1 2>&1 | $(TEST_FILTER)

# Run only e2e tests (filtered by test function name)
test\:e2e:
	@set -o pipefail; go test ./internal/... -v -run E2E -count=1 -p 1 2>&1 | $(TEST_FILTER)
