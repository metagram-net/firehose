.DEFAULT_GOAL := help

.PHONY: help
help: ## List targets in this Makefile
	@awk '\
		BEGIN { FS = ":$$|:[^#]+|:.*?## "; OFS="\t" }; \
		/^[0-9a-zA-Z_-]+?:/ { print $$1, $$2 } \
	' $(MAKEFILE_LIST) \
		| sort --dictionary-order \
		| column --separator $$'\t' --table --table-wrap 2 --output-separator '    '

.PHONY: gen
gen: ## Write generated code to files
	go generate ./...

.PHONY: run
run: ## Run the server
	go run ./cmd/firehose

.PHONY: watch
watch: ## Re-run the server on code changes
	until fd . | entr -dr make run; do echo 'Change detected, restarting server'; done

.PHONY: services
services: ## Start development services
	docker-compose up --detach

.PHONY: services-stop
services-stop: ## Stop development services
	docker-compose stop

.PHONY: services-down
services-down: ## Destroy development services
	docker-compose down

.PHONY: migrate
migrate: ## Run database migrations
	go run ./cmd/drift migrate

.PHONY: lint
lint: ## Run linters
	golangci-lint run

.PHONY: test
test: ## Run Go tests
	go test ./...

.PHONY: test-hurl
test-hurl:
	hurl \
		--progress \
		--summary \
		--output /dev/null \
		--variable root_url="$${TEST_API_ROOT}" \
		--variable basic_auth="$$(echo -n "$${TEST_API_USER}:$${TEST_API_KEY}" | base64 --wrap=0)" \
		./tests/hurl/*

.PHONY: licensed
licensed: licensed-cache licensed-check

.PHONY: licensed-check
licensed-check:
	go mod tidy
	licensed status

.PHONY:
licensed-cache:
	go mod tidy
	licensed cache
