COV="./coverage.txt"
FUNC="./test/base/event/btc/w10101/20260101000000/function"
EXEC="./test/base/agent/execution/202601/01000000"
ACTION="./test/base/event/btc/w10101/20260101000000/action"
ARTIFACT="./test/base/event/btc/w10101/20260101000000/artifact"

.PHONY: test
test: ## Run tests
	go test -coverprofile $(COV) ./...

.PHONY: coverage
coverage: ## Run coverage report
	go tool cover -func $(COV)

.PHONY: vet
vet: ## Run code issues detection
	go vet ./...

.PHONY: clean
clean: ## Remove the bin folder and test files leaving the download data files intact
	rm -rf ./bin ./test/data/btcusd_one_min_data.csv.gz $(COV) $(FUNC)/j* $(ACTION)/j* $(ARTIFACT)/b* $(EXEC)/*.tsv

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
