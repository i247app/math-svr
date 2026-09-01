# Every target is a command, not a file. This must stay complete: `deploy`
# collides with the deploy/ directory, so without .PHONY make sees the target
# as already satisfied and `make deploy` becomes a silent no-op.
.PHONY: help tidy build build-ec2-arm build-ec2-amd run linecount \
	login watch-logs deploy deploy-quick deploy-rollback deploy-amd \
	connect-mysql clear-data-local clear-data-ec2 \
	obs-up obs-down obs-logs obs-reset

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

RHOST ?= none

tidy: ## go tidy
	go mod tidy

build: tidy ## build current or local machine
	go build -o dist/mathsvr ./cmd/mathsvr

build-ec2-arm: tidy ## build AWS EC2 ARM64
	GOOS=linux GOARCH=arm64 go build -o dist/mathsvr ./cmd/mathsvr

build-ec2-amd: tidy ## build AWS EC2 AMD64
	GOOS=linux GOARCH=amd64 go build -o dist/mathsvr-amd64 ./cmd/mathsvr

run: tidy ## run current or local machine
	@go run ./cmd/mathsvr

linecount: ## count lines of code
	find internal pkg -name "*.go" | xargs wc -l

login: ## login to remote host
	@./deploy/scripts/login.sh $(RHOST)

watch-logs: ## watch logs on remote host
	@./deploy/scripts/watch-logs.sh $(RHOST)

deploy:
	@./deploy/scripts/deploy.sh $(RHOST)

deploy-quick:
	@./deploy/scripts/deploy.sh $(RHOST) --skip-build

deploy-rollback:
	@./deploy/scripts/deploy.sh $(RHOST) --rollback

deploy-amd:
	@BUILD_ARCH=amd64 ./deploy/scripts/deploy.sh $(RHOST)

connect-mysql: ## connect to remote mysql
	@./deploy/scripts/connect-mysql.sh

clear-data-local: ## wipe LOCAL user data, keep reference data (uses .env DB_*)
	@./deploy/scripts/clear-data.sh local

clear-data-ec2: ## wipe EC2 user data, keep reference data — var: RHOST=ec2|t1|t2|t3|t4
	@./deploy/scripts/clear-data.sh $(RHOST)

obs-up: ## start prometheus + grafana
	@docker compose -f docker/docker-compose.yml up -d
	@echo "Prometheus → http://localhost:9090"
	@echo "Grafana    → http://localhost:3000  (admin / admin)"

obs-down: ## stop the observability stack (keeps volumes)
	@docker compose -f docker/docker-compose.yml down

obs-logs: ## tail prometheus + grafana logs
	@docker compose -f docker/docker-compose.yml logs -f --tail=100

obs-reset: ## stop and wipe volumes (destroys saved dashboards/data)
	@docker compose -f docker/docker-compose.yml down -v

