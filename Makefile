# .PHONY: build run run-dev build-ec2 deploy login remote-deploy test deploy-quick deploy-rollback deploy-amd watch-logs \
# 	perf-seed perf-targets perf-k6 perf-k6-create perf-k6-get perf-k6-me perf-k6-list perf-k6-update \
# 	perf-vegeta perf-vegeta-create perf-vegeta-get perf-vegeta-me perf-vegeta-list perf-vegeta-update \
# 	perf-clean

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

RHOST ?= none

.PHONY: tidy
tidy: ## go tidy
	go mod tidy

.PHONY: build
build: tidy ## build current or local machine
	go build -o dist/mathsvr ./cmd/mathsvr

.PHONY: build-ec2-arm
build-ec2-arm: tidy ## build AWS EC2 ARM64
	GOOS=linux GOARCH=arm64 go build -o dist/mathsvr ./cmd/mathsvr

.PHONY: build-ec2-amd
build-ec2-amd: tidy ## build AWS EC2 AMD64
	GOOS=linux GOARCH=amd64 go build -o dist/mathsvr-amd64 ./cmd/mathsvr

.PHONY: run
run: tidy ## run current or local machine
	@go run ./cmd/mathsvr

.PHONY: run-dev
run-dev: tidy ## run current or local machine with air
	@air --build.cmd "go build -o dist/mathsvr ./cmd/mathsvr" --build.bin "./dist/mathsvr"

.PHONY: linecount
linecount: ## count lines of code
	find internal pkg -name "*.go" | xargs wc -l

.PHONY: login
login: ## login to remote host
	@./bin/login.sh $(RHOST)

.PHONY: watch-logs
watch-logs: ## watch logs on remote host
	@./bin/watch-logs.sh $(RHOST)

# ── New orchestrated deployment ──────────────────────────

# Full deploy: validate → build → prepare → deliver → activate
.PHONY: deploy
deploy:
	@./bin/deploy.sh $(RHOST)

# Deploy without rebuilding (use existing binary in dist/)
.PHONY: deploy-quick
deploy-quick:
	@./bin/deploy.sh $(RHOST) --skip-build

.PHONY: deploy-rollback
deploy-rollback:
	@./bin/deploy.sh $(RHOST) --rollback

.PHONY: deploy-amd
deploy-amd:
	@BUILD_ARCH=amd64 ./bin/deploy $(RHOST)

.PHONY: connect-mysql
connect-mysql: ## connect to remote mysql
	@./bin/connect-mysql.sh

.PHONY: clear-data-local
clear-data-local: ## wipe LOCAL user data, keep reference data (uses .env DB_*)
	@./bin/clear-data.sh local

.PHONY: clear-data-ec2
clear-data-ec2: ## wipe EC2 user data, keep reference data — var: RHOST=ec2|t1|t2|t3|t4
	@./bin/clear-data.sh $(RHOST)

RATE     ?= 100
DURATION ?= 60s

.PHONY: perf-jmeter-local
perf-jmeter-local: ## load test jmeter local
	@jmeter -n -t perf/jmeter/numi_local.jmx -l perf/jmeter/results/out.jtl -e -o perf/jmeter/report

.PHONY: perf-jmeter-ec2
perf-jmeter-ec2: ## load test jmeter ec2
	@jmeter -n -t perf/jmeter/numi_ec2.jmx -l perf/jmeter/results/out.jtl -e -o perf/jmeter/report


# ── Observability stack (Prometheus + Grafana) ───────────

.PHONY: obs-up
obs-up: ## start prometheus + grafana
	@docker compose -f docker/docker-compose.yml up -d
	@echo "Prometheus → http://localhost:9090"
	@echo "Grafana    → http://localhost:3000  (admin / admin)"

.PHONY: obs-down
obs-down: ## stop the observability stack (keeps volumes)
	@docker compose -f docker/docker-compose.yml down

.PHONY: obs-logs
obs-logs: ## tail prometheus + grafana logs
	@docker compose -f docker/docker-compose.yml logs -f --tail=100

.PHONY: obs-reset
obs-reset: ## stop and wipe volumes (destroys saved dashboards/data)
	@docker compose -f docker/docker-compose.yml down -v

