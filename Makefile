# .PHONY: build run run-dev build-ec2 deploy login remote-deploy test deploy-quick deploy-rollback deploy-amd watch-logs \
# 	perf-seed perf-targets perf-k6 perf-k6-create perf-k6-get perf-k6-me perf-k6-list perf-k6-update \
# 	perf-vegeta perf-vegeta-create perf-vegeta-get perf-vegeta-me perf-vegeta-list perf-vegeta-update \
# 	perf-clean

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

.PHONY: deploy-ec2
deploy-ec2: build-ec2-arm ## build and deploy locally on ec2
	@echo "make[$@] TODO build and deploy locally on ec2..."

.PHONY: deploy-ec2-remote
deploy-ec2-remote: build-ec2-arm ## build and deploy from mac to ec2
	./bin/remote-deploy $(RHOST)
	@echo "make[$@] done"

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

# ── Load testing (user module) ────────────────────────────
# All targets read parameters from perf/.env (copy from perf/.env.example).
# Source it before running, e.g.: `source perf/.env && make perf-seed`
#
# Overridable inline: `make perf-vegeta-create RATE=200 DURATION=120s`

RATE     ?= 100
DURATION ?= 60s

# Seed 10k users + aliases + sessions (writes perf/sessions.gob + perf/sessions.csv).
.PHONY: perf-seed
perf-seed: ## seed 10k users + aliases + sessions
	@go run ./perf/seeder

.PHONY: perf-targets
perf-targets: ## generate Vegeta target files from perf/sessions.csv
	@./perf/vegeta/targets-gen.sh

# ── k6 (ramping VU scenarios) ──
.PHONY: perf-k6-create
perf-k6-create: ## k6 create
	@k6 run perf/k6/create.js

.PHONY: perf-k6-get
perf-k6-get: ## k6 get by id
	@k6 run perf/k6/get_by_id.js

.PHONY: perf-k6-me
perf-k6-me: ## k6 me
	@k6 run perf/k6/me.js

.PHONY: perf-k6-list
perf-k6-list: ## k6 list
	@k6 run perf/k6/list.js

.PHONY: perf-k6-update
perf-k6-update: ## k6 update
	@k6 run perf/k6/update.js

# Run every k6 scenario sequentially.
.PHONY: perf-k6
perf-k6: perf-k6-create perf-k6-get perf-k6-me perf-k6-list perf-k6-update

# ── Vegeta (flat-rate attacks) ──
.PHONY: perf-vegeta-create
perf-vegeta-create: ## vegeta create
	@./perf/vegeta/run.sh create $(RATE) $(DURATION)

.PHONY: perf-vegeta-get
perf-vegeta-get: ## vegeta get by id
	@./perf/vegeta/run.sh get_by_id $(RATE) $(DURATION)

.PHONY: perf-vegeta-me
perf-vegeta-me: ## vegeta me
	@./perf/vegeta/run.sh me $(RATE) $(DURATION)

.PHONY: perf-vegeta-list
perf-vegeta-list: ## vegeta list
	@./perf/vegeta/run.sh list $(RATE) $(DURATION)

.PHONY: perf-vegeta-update
perf-vegeta-update: ## vegeta update
	@./perf/vegeta/run.sh update $(RATE) $(DURATION)

# Run every Vegeta attack sequentially at $(RATE) for $(DURATION).
.PHONY: perf-vegeta
perf-vegeta: perf-vegeta-create perf-vegeta-get perf-vegeta-me perf-vegeta-list perf-vegeta-update

.PHONY: perf-clean
perf-clean: ## wipe generated artefacts (does NOT touch DB rows — see perf/README.md for cleanup SQL)
	@rm -rf perf/sessions.gob perf/sessions.csv perf/vegeta/targets perf/runs
	@echo "perf artefacts removed (DB rows untouched)"
