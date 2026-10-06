# Every target is a command, not a file. This must stay complete: `deploy`
# collides with the deploy/ directory, so without .PHONY make sees the target
# as already satisfied and `make deploy` becomes a silent no-op.
.PHONY: help tidy build build-ec2-arm build-ec2-amd run linecount \
	login watch-logs deploy deploy-quick deploy-rollback deploy-amd \
	connect-mysql migrate migrate-create migrate-status migrate-baseline migrate-down seed db-reset \
	clear-data-local clear-data-ec2 \
	obs-up obs-down obs-logs obs-reset

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-17s\033[0m %s\n", $$1, $$2}'

# Deploy / remote target: t1..t4 pick HOST1..HOST4 from .env.ec2-credentials,
# anything else (including this default) falls back to HOST. login, watch-logs
# and connect-mysql always use HOST and ignore RHOST.
RHOST ?= none

tidy: ## go tidy
	go mod tidy

build: tidy ## build current or local machine
	go build -o dist/mathsvr ./cmd/mathsvr

build-ec2-arm: tidy ## build AWS EC2 ARM64 → dist/mathsvr (the file `make deploy-quick` ships)
	GOOS=linux GOARCH=arm64 go build -o dist/mathsvr ./cmd/mathsvr

build-ec2-amd: tidy ## build AWS EC2 AMD64 → dist/mathsvr-amd64 (not shipped; use deploy-amd)
	GOOS=linux GOARCH=amd64 go build -o dist/mathsvr-amd64 ./cmd/mathsvr

run: tidy ## run current or local machine
	@go run ./cmd/mathsvr

linecount: ## count lines of code
	find internal cmd -name "*.go" | xargs wc -l

login: ## login to remote host (HOST in .env.ec2-credentials)
	@./deploy/scripts/login.sh $(RHOST)

watch-logs: ## watch logs on remote host (HOST in .env.ec2-credentials)
	@./deploy/scripts/watch-logs.sh $(RHOST)

deploy: ## full deploy: validate → build (arm64) → prepare → deliver → activate — RHOST=t1|t2|t3|t4
	@./deploy/scripts/deploy.sh $(RHOST)

deploy-quick: ## deploy the existing dist/mathsvr without rebuilding — RHOST=t1|t2|t3|t4
	@./deploy/scripts/deploy.sh $(RHOST) --skip-build

deploy-rollback: ## restore the previous binary on the host — RHOST=t1|t2|t3|t4
	@./deploy/scripts/deploy.sh $(RHOST) --rollback

deploy-amd: ## full deploy with an AMD64 build — RHOST=t1|t2|t3|t4
	@BUILD_ARCH=amd64 ./deploy/scripts/deploy.sh $(RHOST)

connect-mysql: ## connect to remote mysql (HOST in .env.ec2-credentials)
	@./deploy/scripts/connect-mysql.sh

migrate: ## apply pending migrations/up/*.sql to LOCAL db (uses .env DB_*)
	@./deploy/scripts/migrate.sh up

migrate-create: ## create new migration pair — var: NAME=ma_foos
	@./deploy/scripts/create_migration.sh $(NAME)

migrate-status: ## list applied / pending migrations on LOCAL db, change nothing
	@./deploy/scripts/migrate.sh status

migrate-down: ## DESTRUCTIVE: run migrations/down/ for every applied version — drops all tables
	@./deploy/scripts/migrate.sh down

migrate-baseline: ## mark all migrations as applied WITHOUT running them (one-off, existing LOCAL db)
	@./deploy/scripts/migrate.sh baseline

seed: ## run migrations/seed/*.sql (ma_seqs, programs, grades, semesters) — idempotent
	@./deploy/scripts/migrate.sh seed

db-reset: migrate-down migrate seed ## DESTRUCTIVE: down → up → seed, a clean LOCAL database

clear-data-local: ## wipe LOCAL user data, keep reference data (uses .env DB_*)
	@./deploy/scripts/clear-data.sh local

clear-data-ec2: ## wipe EC2 user data, keep reference data — var: RHOST=ec2|t1|t2|t3|t4
	@./deploy/scripts/clear-data.sh $(RHOST)

obs-up: ## start the observability stack: prometheus, loki, tempo, alloy, grafana
	@docker compose -f docker/docker-compose.yml up -d
	@echo "Grafana    → http://localhost:3000  (anonymous admin locally; or admin / admin)"
	@echo "Prometheus → http://localhost:9090"
	@echo "Loki       → http://localhost:3100   Tempo → http://localhost:3200"
	@echo "Pair with OBS_TRACING_ENABLED=true, LOG_FILE=./logs/app.log, LOG_FILE_FORMAT=json in .env (docker/README.md)"

obs-down: ## stop the observability stack (keeps volumes)
	@docker compose -f docker/docker-compose.yml down

obs-logs: ## tail the observability stack logs
	@docker compose -f docker/docker-compose.yml logs -f --tail=100

obs-reset: ## stop and wipe volumes (destroys saved dashboards/data)
	@docker compose -f docker/docker-compose.yml down -v
