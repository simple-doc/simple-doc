.PHONY: db-up db-down db-restart db-logs db-psql db-reset seed build run dev build-docker

db-up:
	docker compose -p simple-doc up -d postgres

db-down:
	docker compose -p simple-doc down

db-restart:
	docker compose -p simple-doc restart postgres

db-logs:
	docker compose -p simple-doc logs -f postgres

db-psql:
	docker compose -p simple-doc exec postgres psql -U postgres

db-reset:
	docker compose -p simple-doc down -v
	docker compose -p simple-doc up -d postgres

seed:
	go run cmd/seed/main.go

build:
	go build -o server cmd/server/main.go

run:
	go run cmd/server/main.go

dev: db-up
	@echo "Waiting for PostgreSQL..."
	@sleep 2
	@$(MAKE) seed
	@$(MAKE) run

build-docker:
	docker build -t simple-doc:latest .
