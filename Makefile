.PHONY: help install update build test race vet staticcheck fmt fmt-check qa example-invoice example-info example-history example-static-wallet example-webhook example-balance example-payout example-payout-info example-services example-rates webhook-inspect docker-build docker-shell docker-webhook docker-qa clean

GO          ?= go
STATICCHECK ?= staticcheck

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

install: ## Download module dependencies
	$(GO) mod download

update: ## Update module dependencies
	$(GO) get -u ./... && $(GO) mod tidy

build: ## Build the heleket-webhook-inspect CLI to bin/
	mkdir -p bin
	$(GO) build -o bin/heleket-webhook-inspect ./cmd/heleket-webhook-inspect

test: ## Run all unit tests
	$(GO) test ./...

race: ## Run all unit tests with the race detector
	$(GO) test -race ./...

vet: ## Run go vet across the module
	$(GO) vet ./...

staticcheck: ## Run staticcheck (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)
	$(STATICCHECK) ./...

fmt: ## Apply gofmt to all files
	$(GO) fmt ./...

fmt-check: ## Fail if any file would be reformatted
	@diff -u <(echo -n) <(gofmt -l .) || (echo "gofmt found issues; run 'make fmt'" && exit 1)

qa: vet staticcheck race fmt-check ## Run all quality gates (vet + staticcheck + race + fmt)

example-invoice: ## Create a test invoice via the API
	$(GO) run ./examples/create_invoice

example-info: ## Look up payment info (pass UUID env var)
	$(GO) run ./examples/get_payment_info $(UUID)

example-history: ## List recent payments
	$(GO) run ./examples/list_payment_history

example-static-wallet: ## Create a static (top-up) wallet
	$(GO) run ./examples/create_static_wallet

example-webhook: ## Run the webhook handler on http://localhost:8000
	$(GO) run ./examples/handle_webhook

example-balance: ## Show merchant + personal balances
	$(GO) run ./examples/get_balance

example-payout: ## Create a payout (pass AMOUNT and ADDRESS env vars)
	$(GO) run ./examples/create_payout $(AMOUNT) $(ADDRESS)

example-payout-info: ## Look up a payout (pass UUID env var)
	$(GO) run ./examples/get_payout_info $(UUID)

example-services: ## List payment or payout services (KIND=payment|payout)
	$(GO) run ./examples/list_services $(KIND)

example-rates: ## List exchange rates (CURRENCY=USD by default)
	$(GO) run ./examples/exchange_rates $(CURRENCY)

webhook-inspect: build ## Pipe a webhook payload into the inspector CLI (pass KEY env var)
	./bin/heleket-webhook-inspect --key=$(KEY)

docker-build: ## Build the dev Docker image
	docker compose build

docker-shell: ## Open a shell in the dev container
	docker compose run --rm cli sh

docker-webhook: ## Run the webhook handler in Docker on port 8000
	docker compose up webhook

docker-qa: ## Run the QA pipeline in Docker
	docker compose run --rm qa

clean: ## Remove generated artifacts
	rm -rf bin/ coverage.out coverage.html
