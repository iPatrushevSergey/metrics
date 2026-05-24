.PHONY: build build-server build-agent run-server run-agent test test-unit test-contract test-component test-integration test-e2e test-all cover cover-unit migrate generate-mocks generate-proto

APP_DIR := app
PROTO_DIR := $(APP_DIR)/internal/pkg/grpc/metrics
PROTO_FILE := $(PROTO_DIR)/metrics.proto

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

test-component:
	cd $(APP_DIR) && go test ./tests/component/...

test-integration:
	cd $(APP_DIR) && go test -tags=integration ./internal/pkg/adapters/repository/postgres/...

test-e2e:
	cd $(APP_DIR) && go test ./tests/e2e/...

test-all: test-unit test-contract test-component test-integration test-e2e

# Coverage: unit + integration + e2e + component + contract tests.
cover:
	cd $(APP_DIR) && go test -tags=integration -coverpkg=./... ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# Coverage: unit tests only.
cover-unit:
	cd $(APP_DIR) && go test -coverpkg=./... ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# DATABASE_DSN must be set, e.g. postgres://user:pass@localhost:5432/metrics?sslmode=disable
migrate:
	cd $(APP_DIR) && go run ./cmd/migrate -d "$(DATABASE_DSN)"

# Regenerate gomock stubs (committed under port/mocks; run after port interface changes).
MOCKGEN := go run go.uber.org/mock/mockgen@v0.6.0
SERVER_PORT := internal/server/metrics/application/port
AGENT_PORT := internal/agent/collector/application/port

generate-proto:
	protoc \
		--go_out=. --go_opt=module=github.com/iPatrushevSergey/metrics \
		--go-grpc_out=. --go-grpc_opt=module=github.com/iPatrushevSergey/metrics \
		$(PROTO_FILE)

generate-mocks:
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/metrics_repository.go -destination=$(SERVER_PORT)/mocks/mock_metrics_repository.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/metric_file_repository.go -destination=$(SERVER_PORT)/mocks/mock_metric_file_repository.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/audit_publisher.go -destination=$(SERVER_PORT)/mocks/mock_audit_publisher.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/audit_gateway.go -destination=$(SERVER_PORT)/mocks/mock_audit_gateway.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/audit_file_repository.go -destination=$(SERVER_PORT)/mocks/mock_audit_file_repository.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/transactor.go -destination=$(SERVER_PORT)/mocks/mock_transactor.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(SERVER_PORT)/logger.go -destination=$(SERVER_PORT)/mocks/mock_logger.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(AGENT_PORT)/metrics_repository.go -destination=$(AGENT_PORT)/mocks/mock_metrics_repository.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(AGENT_PORT)/metrics_gateway.go -destination=$(AGENT_PORT)/mocks/mock_metrics_gateway.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(AGENT_PORT)/metrics_sampler.go -destination=$(AGENT_PORT)/mocks/mock_metrics_sampler.go -package=mocks
	cd $(APP_DIR) && $(MOCKGEN) -source=$(AGENT_PORT)/logger.go -destination=$(AGENT_PORT)/mocks/mock_logger.go -package=mocks
