.PHONY: deps build test test-integration lint proto sqlc oapi generate up down up-jobs down-jobs migrate-user migrate-payment migrate-notification migrate-card tools reconcile-payment reconcile-card list-payment-breaks list-card-breaks saga-watchdog list-saga-alerts

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

up:
	docker compose -f deployments/docker-compose.yml up -d

down:
	docker compose -f deployments/docker-compose.yml down

up-jobs:
	docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.jobs.yml up -d --build

down-jobs:
	docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.jobs.yml down

migrate-user:
	cd services/user && go run ./cmd/migrate

migrate-payment:
	cd services/payment && go run ./cmd/migrate

migrate-notification:
	cd services/notification && go run ./cmd/migrate

migrate-card:
	cd services/card && go run ./cmd/migrate

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