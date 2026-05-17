.PHONY: build run run-dev build-ec2 deploy login remote-deploy test deploy-quick deploy-rollback deploy-amd watch-logs 

RHOST ?= none

tidy:
	go mod tidy

# build current or local machine
build: tidy
	go build -o dist/mathsvr ./cmd/mathsvr

# build AWS EC2 ARM64
build-ec2-arm: tidy
	GOOS=linux GOARCH=arm64 go build -o dist/mathsvr ./cmd/mathsvr

# build AWS EC2 AMD64
build-ec2-amd: tidy
	GOOS=linux GOARCH=amd64 go build -o dist/mathsvr-amd64 ./cmd/mathsvr

deploy-ec2: build-ec2-arm
	@echo "make[$@] TODO build and deploy locally on ec2..."

# build local, deploy
deploy-ec2-remote: build-ec2-arm
	@echo "make[$@] build and deploy from mac to ec2..."
	./bin/remote-deploy $(RHOST)
	@echo "make[$@] done"

run: tidy
	@go run ./cmd/mathsvr

run-dev: tidy
	@air --build.cmd "go build -o dist/mathsvr ./cmd/mathsvr" --build.bin "./dist/mathsvr"

check-nil: tidy
	go install go.uber.org/nilaway/cmd/nilaway@latest   
	nilaway -include-pkgs="monex.com/monex" ./...

linecount:
	find internal pkg -name "*.go" | xargs wc -l

login:
	@./bin/login.sh $(RHOST)

watch-logs:
	@./bin/watch-logs.sh $(RHOST)

# ── New orchestrated deployment ──────────────────────────

# Full deploy: validate → build → prepare → deliver → activate
deploy:
	@./bin/deploy.sh $(RHOST)

# Deploy without rebuilding (use existing binary in dist/)
deploy-quick:
	@./bin/deploy.sh $(RHOST) --skip-build

# Rollback to previous binary
deploy-rollback:
	@./bin/deploy.sh $(RHOST) --rollback

# Build for AMD64 and deploy
deploy-amd:
	@BUILD_ARCH=amd64 ./bin/deploy $(RHOST)
