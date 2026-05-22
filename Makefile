.PHONY: build build-server build-agent run-server run-agent test test-unit test-contract test-integration test-e2e test-all cover migrate

APP_DIR := app

# go_json: Gin uses goccy/go-json instead of encoding/json for binding/rendering
build: build-server build-agent

build-server:
	cd $(APP_DIR) && go build -tags=go_json -o bin/server ./cmd/server

build-agent:
	cd $(APP_DIR) && go build -tags=go_json -o bin/agent ./cmd/agent

run-server:
	cd $(APP_DIR) && go run -tags=go_json ./cmd/server

run-agent:
	cd $(APP_DIR) && go run -tags=go_json ./cmd/agent

test:
	$(MAKE) test-unit

test-unit:
	cd $(APP_DIR) && go test ./internal/... ./cmd/...

test-contract:
	cd $(APP_DIR) && go test ./tests/contract/...

test-integration:
	cd $(APP_DIR) && go test -tags=integration ./internal/pkg/adapters/repository/postgres/...

test-e2e:
	cd $(APP_DIR) && go test -tags=integration ./tests/e2e/...

test-all: test-unit test-contract test-integration test-e2e

cover:
	cd $(APP_DIR) && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# DATABASE_DSN must be set, e.g. postgres://user:pass@localhost:5432/metrics?sslmode=disable
migrate:
	cd $(APP_DIR) && go run ./cmd/migrate -d "$(DATABASE_DSN)"
