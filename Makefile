-include .env
export

.PHONY: build run docker-up docker-down migrate-up migrate-down test lint help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the API server"
	@echo "  run           Run the API server"
	@echo "  docker-up     Start infrastructure (Postgres)"
	@echo "  docker-down   Stop infrastructure"
	@echo "  migrate-up    Apply all migrations"
	@echo "  migrate-down  Rollback last migration"
	@echo "  test          Run all tests"
	@echo "  lint          Run golangci-lint"

build:
	go build -o tmp/api ./cmd/api

run: build
	./tmp/api

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

test:
	go test -v ./...

lint:
	golangci-lint run
