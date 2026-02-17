APP_NAME=url-shortener

# Load environment variables from .env file
include .env
export

.PHONY: dev prod build up down logs db deploy clean help


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
	@echo "  deploy         - Deploy to production"