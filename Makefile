.PHONY: build run list-models tidy migrate-create login

## run: Run the app.
run:
	@go run ./cmd/server

list-models:
	@go run ./cmd/list-models

tidy:
	go mod tidy

# build current or local machine
build: tidy
	go build -o dist/server ./cmd/server

# Get list of .sql files from the up/ directory, sorted from old to new
MIGRATE_UP_FILES := $(shell ls migrations/up/*.sql | sort)

# Get list of .sql files from the down/ directory, sorted from new to old
MIGRATE_DOWN_FILES := $(shell ls migrations/down/*.sql | sort -r)

## migrate-up: Run all pending migrations to the database.
migrate-up:
	@echo "🚀 Starting UP migrations from /migrations/up..."
	@if [ -z "$(MIGRATE_UP_FILES)" ]; then \
		echo "No UP migration files found in migrations/up."; \
	else \
		for file in $(MIGRATE_UP_FILES); do \
			echo "--> Running UP: $$file"; \
			./bin/run_migration_file.sh $$file; \
		done; \
	fi
	@echo "✅ All UP migrations completed."

## migrate-down: Run all migrations to the database.
migrate-down:
	@echo "⏪ Starting DOWN migrations from /migrations/down..."
	@if [ -z "$(MIGRATE_DOWN_FILES)" ]; then \
		echo "No DOWN migration files found in migrations/down."; \
	else \
		for file in $(MIGRATE_DOWN_FILES); do \
			echo "--> Running DOWN: $$file"; \
			./bin/run_migration_file.sh $$file; \
		done; \
	fi
	@echo "✅ All DOWN migrations completed."


## create-migration NAME=<name>: Create a new migration file.
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration NAME=<migration_name>"; \
		exit 1; \
	fi
	./bin/create_migration.sh $(NAME)

## list-models: List all available Google AI models
list-models:
	@if [ ! -f ./bin/list-models ]; then \
		echo "Building list-models utility..."; \
		go build -o bin/list-models ./cmd/list-models; \
	fi
	@./bin/list-models


## login: Login to AWS