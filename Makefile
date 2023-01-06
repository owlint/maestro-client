PROJECT_NAME := $(if $(PROJECT_NAME),$(PROJECT_NAME),maestro-client)

.PHONY: all up bare_test down test generate

up:
	docker-compose -p ${PROJECT_NAME} up -d

bare_test:
	REDIS_PORT=6379 REDIS_HOST=localhost go test ./...

down:
	-docker-compose -p ${PROJECT_NAME} down

test: down up bare_test down

generate:
	go generate ./...

all: generate test run
