APP_NAME=url-shortener

.PHONY: dev prod build up down logs db


dev:
	go run ./main.go

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

## View logs
logs:
	docker compose logs -f

## Connect to DB
db:
	docker exec -it postgres_db psql -U postgres -d go_app_db