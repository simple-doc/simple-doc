.PHONY: db-up db-down db-restart db-logs db-psql db-reset migrate seed seed-minimal build run run-loop dev build-docker run-docker export import

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

migrate:
	go run cmd/migrate/main.go

seed:
	go run cmd/seed/main.go

seed-minimal:
	go run cmd/seed/main.go -minimal

build:
	go build -o server cmd/server/main.go

run:
	air

run-loop:
	@while true; do \
		go run cmd/server/main.go; \
		echo "Server exited. Restarting..."; \
	done

dev: db-up
	@echo "Waiting for PostgreSQL..."
	@sleep 2
	@$(MAKE) seed
	@$(MAKE) run

export:
	go run cmd/portability/main.go export -o export-$(shell date +%Y%m%d-%H%M%S).json

import:
	@test -n "$(FILE)" || (echo "Usage: make import FILE=backup.json" && exit 1)
	go run cmd/portability/main.go import -i $(FILE)

build-docker:
	docker build -t simple-doc:latest .

run-docker:
	docker compose -p simple-doc --profile docker up -d --build
