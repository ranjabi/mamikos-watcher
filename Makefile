.PHONY: help test test-integration fmt fmt-fix vet check run-once

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-18s %s\n", $$1, $$2}'

test: ## Run unit tests (skips live tests)
	go test ./...

test-integration: ## Run all tests including live tests against mamikos.com
	INTEGRATION=1 go test -v ./...

fmt: ## Fail if any Go file is not gofmt-formatted
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "needs gofmt (run: make fmt-fix):"; echo "$$out"; exit 1; fi

fmt-fix: ## Format all Go files in place
	gofmt -w .

vet: ## Run go vet
	go vet ./...

check: fmt vet test ## Run fmt, vet and unit tests

run-once: ## Check every URL in .env one time and exit
	@test -f .env || { echo ".env not found (run: cp .env.example .env)"; exit 1; }
	set -a && . ./.env && set +a && go run ./cmd/watcher -once
