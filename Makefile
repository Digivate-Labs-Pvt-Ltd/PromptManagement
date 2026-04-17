-include .env
export

.PHONY: run docker-up docker-down migrate-up migrate-down test help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  run           Run the API server"
	@echo "  docker-up     Start infrastructure (Postgres)"
	@echo "  docker-down   Stop infrastructure"
	@echo "  migrate-up    Apply all migrations"
	@echo "  migrate-down  Rollback last migration"
	@echo "  test          Run all tests"

run:
	go run cmd/api/main.go

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
