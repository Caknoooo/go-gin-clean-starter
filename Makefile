# Help command
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  help          Show this help message"
	@echo "  dep           Run go mod tidy"
	@echo "  run           Run the application"
	@echo "  build         Build the application"
	@echo "  run-build     Build and run the application"
	@echo "  test          Run tests"
	@echo "  init-docker   Initialize and start docker containers"
	@echo "  up            Start docker containers"
	@echo "  down          Stop docker containers"
	@echo "  logs          Show container logs"
	@echo "  container-postgres  Access PostgreSQL container shell"
	@echo "  create-db     Create database"
	@echo "  init-uuid     Initialize UUID extension"
	@echo "  container-go  Access Go application container shell"
	@echo "  migrate       Run database migrations"
	@echo "  seed          Run database seeds"
	@echo "  migrate-seed  Run both migrations and seeds"
	@echo "  go-tidy       Run go mod tidy in container"
	@echo "  swagger       Generate Swagger documentation"
	@echo "  clean-swagger Clean generated Swagger documentation"

# Import .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables
CONTAINER_NAME=$(APP_NAME)-app
POSTGRES_CONTAINER_NAME=$(APP_NAME)-db

# Commands
dep:
	go mod tidy

run:
	go run main.go

build:
	go build -o main main.go

run-build: build
	./main

test:
	APP_ENV=test env $(cat .env.test | xargs) go test -v ./...

init-docker:
	docker compose up -d --build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Postgres commands
container-postgres:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh

create-db:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh -c "createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}"

init-uuid:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh -c "psql -U ${DB_USER} -d ${DB_NAME} -c 'CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";'"

# Docker commands
container-go:
	docker exec -it ${CONTAINER_NAME} /bin/sh

migrate:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run main.go --migrate"

seed:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run main.go --seed"

migrate-seed:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run main.go --migrate --seed"

go-tidy:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go mod tidy"

swagger:
	swag init

clean-swagger:
	rm -rf docs/