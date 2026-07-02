.PHONY: deps build test test-integration lint proto sqlc oapi generate up down up-all down-all up-ghcr down-ghcr up-jobs down-jobs migrate migrate-user migrate-payment migrate-notification migrate-card vault-init tools reconcile-payment reconcile-card list-payment-breaks list-card-breaks saga-watchdog list-saga-alerts aml-export event-catalog grpc-mtls-certs helm-lint helm-template

HELM_CHART := deploy/helm/neobank

COMPOSE_INFRA := docker compose -f deployments/docker-compose.yml
COMPOSE_ALL   := docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.services.yml
COMPOSE_GHCR  := $(COMPOSE_ALL) -f deployments/docker-compose.images.yml
COMPOSE_JOBS  := $(COMPOSE_ALL) -f deployments/docker-compose.jobs.yml
GIT_SHA       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BUILD_DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

OAPI_CODEGEN ?= oapi-codegen
SQLC ?= sqlc

tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

deps:
	cd pkg && go mod tidy
	cd services/user && go mod tidy
	cd services/payment && go mod tidy
	cd services/gateway && go mod tidy
	cd services/notification && go mod tidy
	cd services/card && go mod tidy

proto:
	cd proto && buf dep update && buf generate

sqlc:
	cd services/user && $(SQLC) generate
	cd services/payment && $(SQLC) generate
	cd services/notification && $(SQLC) generate
	cd services/card && $(SQLC) generate

oapi:
	cd services/user && $(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml
	cd services/payment && $(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml
	cd services/gateway && $(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml
	cd services/notification && $(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml
	cd services/card && $(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml

generate: proto sqlc oapi

grpc-mtls-certs:
	chmod +x deployments/grpc-mtls/gen-certs.sh
	./deployments/grpc-mtls/gen-certs.sh

build: generate
	go build -o bin/user ./services/user/cmd/server
	go build -o bin/payment ./services/payment/cmd/server
	go build -o bin/gateway ./services/gateway/cmd/server
	go build -o bin/notification ./services/notification/cmd/server
	go build -o bin/card ./services/card/cmd/server
	go build -o bin/payment-reconcile ./services/payment/cmd/reconcile
	go build -o bin/card-reconcile ./services/card/cmd/reconcile
	go build -o bin/payment-resolve-break ./services/payment/cmd/resolve-break
	go build -o bin/card-resolve-break ./services/card/cmd/resolve-break
	go build -o bin/saga-watchdog ./tools/saga-watchdog
	go build -o bin/payment-aml-export ./services/payment/cmd/aml-export
	go build -o bin/event-catalog ./tools/event-catalog

test:
	cd pkg && go test ./...

test-integration:
	cd tests/integration && go test -v -count=1 -timeout 15m ./...
	cd services/user && go test ./...
	cd services/payment && go test ./...
	cd services/gateway && go test ./...
	cd services/notification && go test ./...
	cd services/card && go test ./...

lint:
	cd pkg && golangci-lint run --config=../.golangci.yml ./...
	cd services/gateway && golangci-lint run --config=../../.golangci.yml ./...
	cd services/user && golangci-lint run --config=../../.golangci.yml ./...
	cd services/payment && golangci-lint run --config=../../.golangci.yml ./...
	cd services/card && golangci-lint run --config=../../.golangci.yml ./...
	cd services/notification && golangci-lint run --config=../../.golangci.yml ./...
	cd tests/integration && golangci-lint run --config=../../.golangci.yml ./...
	cd tools/saga-watchdog && golangci-lint run --config=../../.golangci.yml ./...
	cd tools/event-catalog && golangci-lint run --config=../../.golangci.yml ./...

up:
	$(COMPOSE_INFRA) up -d

down:
	$(COMPOSE_INFRA) down

up-all:
	GIT_SHA=$(GIT_SHA) BUILD_DATE=$(BUILD_DATE) $(COMPOSE_ALL) up -d --build

down-all:
	$(COMPOSE_ALL) down

up-ghcr:
	$(COMPOSE_GHCR) up -d --no-build --pull always

down-ghcr:
	$(COMPOSE_GHCR) down

up-jobs:
	GIT_SHA=$(GIT_SHA) BUILD_DATE=$(BUILD_DATE) $(COMPOSE_JOBS) up -d --build reconcile-jobs

down-jobs:
	$(COMPOSE_JOBS) stop reconcile-jobs

migrate: migrate-user migrate-payment migrate-notification migrate-card

migrate-user:
	cd services/user && go run ./cmd/migrate

migrate-payment:
	cd services/payment && go run ./cmd/migrate

migrate-notification:
	cd services/notification && go run ./cmd/migrate

migrate-card:
	cd services/card && go run ./cmd/migrate

vault-init:
	./deployments/vault-init.sh

reconcile-payment:
	cd services/payment && go run ./cmd/reconcile

reconcile-card:
	cd services/card && go run ./cmd/reconcile

list-payment-breaks:
	cd services/payment && go run ./cmd/resolve-break -list

list-card-breaks:
	cd services/card && go run ./cmd/resolve-break -list

saga-watchdog:
	cd tools/saga-watchdog && go run .

list-saga-alerts:
	cd tools/saga-watchdog && go run . -list

aml-export:
	cd services/payment && go run ./cmd/aml-export

event-catalog:
	cd tools/event-catalog && go run .

helm-lint:
	helm lint $(HELM_CHART) -f $(HELM_CHART)/values-staging.yaml

helm-template:
	helm template neobank $(HELM_CHART) -f $(HELM_CHART)/values-staging.yaml \
		--set secrets.create=true \
		--set config.databaseURL=postgres://neobank:neobank@postgres:5432/neobank?sslmode=disable \
		--set config.redisURL=redis://redis:6379/0 \
		--set config.jwtSecret=local-dev-secret \
		--set config.kafkaBrokers=redpanda:9092