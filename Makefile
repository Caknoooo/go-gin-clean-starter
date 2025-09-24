# Import .env file
ifneq (,$(wildcard ./.env))
		include .env
		export $(shell sed 's/=.*//' .env)
endif

# Variables
CONTAINER_NAME=${APP_NAME}-app
POSTGRES_CONTAINER_NAME=${APP_NAME}-db

# Commands
dep: 
	go mod tidy

run: 
	go run cmd/main.go

build: 
	go build -o main cmd/main.go

run-build: build
	./main

test:
	go test -v ./tests

test-auth:
	go test -v ./modules/auth/tests/...

test-user:
	go test -v ./modules/user/tests/...

test-all:
	go test -v ./modules/.../tests/...

test-coverage:
	go test -v -coverprofile=coverage.out ./modules/.../tests/...
	go tool cover -html=coverage.out

module:
	@if [ -z "$(name)" ]; then echo "Usage: make module name=<module_name>"; exit 1; fi
	@./create_module.sh $(name)

# Local commands (without docker)
migrate-local:
	go run cmd/main.go --migrate

seed-local: 
	go run cmd/main.go --seed

migrate-seed-local: 
	go run cmd/main.go --migrate --seed

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
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate"

seed: 
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --seed"

migrate-seed: 
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate --seed"

go-tidy:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go mod tidy"