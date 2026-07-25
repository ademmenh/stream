.PHONY: help test test\:unit test\:integration test\:e2e lint format generate clean migrate\:create migrate\:apply

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
	@echo -e "  Code Generation"
	@echo -e "    $(CYAN)make generate$(RESET)           Regenerate Ent code from schema"
	@echo -e "    $(CYAN)make migrate:create$(RESET)     Generate Atlas migration from schema"
	@echo -e "    $(CYAN)make migrate:apply$(RESET)      Apply pending Atlas migrations"
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
	$(ATLAS) migrate diff --dir "file://migrations" --to "ent://internal/shared/infrastructure/ent/schema" --dev-url "postgres://postgres:postgres@localhost:5432/atlas_dev?sslmode=disable"

migrate\:apply:
	$(ATLAS) migrate apply --dir "file://migrations" --url "postgres://postgres:postgres@localhost:5432/starter?sslmode=disable" --allow-dirty

# ── Tests ──────────────────────────────────────────────────────────────────────
TEST_FILTER = grep -v '^? ' | stdbuf -oL sed -E 's/(--- PASS:.*)/\x1b[32m\1\x1b[0m/g; s/(--- FAIL:.*)/\x1b[31m\1\x1b[0m/g; s/(--- SKIP:.*)/\x1b[33m\1\x1b[0m/g; s/^(ok .*)/\x1b[32m\1\x1b[0m/g; s/^(FAIL .*)/\x1b[31m\1\x1b[0m/g'

# Run all tests (unit + integration + e2e) — requires postgres + minio
test:
	@set -o pipefail; go test ./... -v -count=1 -p 1 2>&1 | $(TEST_FILTER)

# Run only unit tests (skips integration & e2e via testing.Short)
test\:unit:
	@set -o pipefail; go test ./... -v -count=1 -short 2>&1 | $(TEST_FILTER)

# Run only integration tests (requires postgres + minio)
test\:integration:
	@set -o pipefail; go test ./internal/... -v -run Integration -count=1 -p 1 2>&1 | $(TEST_FILTER)

# Run only e2e tests (requires postgres + minio)
test\:e2e:
	@set -o pipefail; go test ./internal/... -v -run E2E -count=1 -p 1 2>&1 | $(TEST_FILTER)
