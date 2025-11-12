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

# Commands (without docker)
migrate:
	go run cmd/main.go --migrate:run

migrate-rollback:
	go run cmd/main.go --migrate:rollback

migrate-rollback-batch:
	@if [ -z "$(batch)" ]; then echo "Usage: make migrate-rollback-batch batch=<batch_number>"; exit 1; fi
	go run cmd/main.go --migrate:rollback $(batch)

migrate-rollback-all:
	go run cmd/main.go --migrate:rollback:all

migrate-status:
	go run cmd/main.go --migrate:status

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	go run cmd/main.go --migrate:create:$(name)

seed: 
	go run cmd/main.go --seed

migrate-seed: 
	go run cmd/main.go --migrate:run --seed

# Postgres commands
container-postgres:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh

create-db:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh -c "createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}"

init-uuid:
	docker exec -it ${POSTGRES_CONTAINER_NAME} /bin/sh -c "psql -U ${DB_USER} -d ${DB_NAME} -c 'CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";'"

# Docker commands
init-docker:
	docker compose up -d --build

up: 
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

container-go:
	docker exec -it ${CONTAINER_NAME} /bin/sh

migrate-docker:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:run"

migrate-rollback-docker:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:rollback"

migrate-rollback-batch-docker:
	@if [ -z "$(batch)" ]; then echo "Usage: make migrate-rollback-batch-docker batch=<batch_number>"; exit 1; fi
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:rollback $(batch)"

migrate-rollback-all-docker:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:rollback:all"

migrate-status-docker:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:status"

migrate-create-docker:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create-docker name=<migration_name>"; exit 1; fi
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:create:$(name)"

seed-docker: 
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --seed"

migrate-seed-docker: 
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go run cmd/main.go --migrate:run --seed"

go-tidy-docker:
	docker exec -it ${CONTAINER_NAME} /bin/sh -c "go mod tidy"