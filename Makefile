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
	go run .

.PHONY: watch
watch: ## Re-run the server on code changes
	until fd . | entr -dr make run; do echo 'Change detected, restarting server'; done

.PHONY: services
services:
	docker-compose up --detach

.PHONY: services-stop
services-stop:
	docker-compose stop

.PHONY: services-down
services-down:
	docker-compose down

.PHONY: migrate
migrate:
	go run ./cmd/drift migrate

.PHONY: lint
lint:
	golangci-lint run
