.PHONY: help
help: ## Display this help text
	@grep -hE '^[A-Za-z0-9_ \-]*?:.*##.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Installs dependencies
	@./build/deps.sh

.PHONY: deps-upgrade
deps-upgrade: ## Installs/upgrades all dependencies
	@./build/deps-upgrade.sh

.PHONY: lint
lint: ## Runs linting
	@./build/lint.sh

.PHONY: test
test: ## Run all tests
	@./build/test.sh

.PHONY: run
run: ## Runs the application (containerised)
	@docker-compose up coinbase-vwap

.PHONY: quality
quality: lint test ## Runs all quality checks
	@echo Done