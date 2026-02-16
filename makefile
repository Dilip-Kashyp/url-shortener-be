APP_NAME=url-shortener

# Load environment variables from .env file
include .env
export

.PHONY: dev prod build up down logs db test test-coverage test-report deploy clean help


dev:
	go run ./cmd/server/main.go

prod:
	docker compose up --build

## Build only
build:
	docker build -t $(APP_NAME) .

## Start containers
up:
	docker compose up -d

## Stop containers
down:
	docker compose down

## Stop containers and remove volumes
clean:
	docker compose down -v
	docker image prune -f

## View logs
logs:
	docker compose logs -f

## Connect to DB (uses environment variables)
db:
	docker exec -it postgres_db psql -U $(DB_USER) -d $(DB_NAME)

## Run all tests
test:
	go test -v ./...

## Run tests with coverage
test-coverage:
	go test -v -cover ./...

## Run tests with coverage report
test-report:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Production deploy (used by CI/CD)
deploy:
	git pull origin main
	docker compose down
	docker compose up -d --build
	docker image prune -f

## Show help
help:
	@echo "Available targets:"
	@echo "  dev            - Run app locally with Go"
	@echo "  prod           - Run with Docker Compose (build)"
	@echo "  build          - Build Docker image"
	@echo "  up             - Start containers in background"
	@echo "  down           - Stop containers"
	@echo "  clean          - Stop containers and remove volumes"
	@echo "  logs           - View container logs"
	@echo "  db             - Connect to PostgreSQL"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-report    - Generate HTML coverage report"
	@echo "  deploy         - Deploy to production"