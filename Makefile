include .env

run:
	docker-compose -f docker-compose.yml -p $(PROJECT_NAME) up -d

down:
	docker-compose -f docker-compose.yml -p $(PROJECT_NAME) down --remove-orphans -v

server:
	go run cmd/audit-logging-api/main.go

migrate:
	docker-compose -f docker-compose-db-tools.yml run --rm flyway migrate
	docker compose -f docker-compose.yml exec -T timescaledb \
	pg_dump -U $(POSTGRES_USER) -d $(POSTGRES_DB) \
	--schema-only \
	--no-owner \
	--no-privileges \
	--exclude-table=flyway_schema_history \
	> migrations/structure.sql

GEN_DIR := gen/specs
codegen-unify:
	docker-compose -f docker-compose-tools.yml run --rm openapi-generator-cli generate -g openapi-yaml -i /api/api_service.v1.yaml -o /api/$(GEN_DIR)

codegen: codegen-unify
	# api
	mkdir -p ./internal/adapter/http/gen/api
	docker-compose -f docker-compose-tools.yml run --rm oapi-codegen\
		-generate "types" -package api_service /api/gen/specs/openapi/openapi.yaml > ./internal/adapter/http/gen/api/service.types.go
	docker-compose -f docker-compose-tools.yml run --rm oapi-codegen\
		-generate "gin-server,spec" -package api_service /api/gen/specs/openapi/openapi.yaml > ./internal/adapter/http/gen/api/service.server.go