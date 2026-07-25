.PHONY: help build run clean vet

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o blog .

run: build ## Build and run the server
	PORT=8080 DB_PATH=blog.db ./blog

clean: ## Remove the binary and database
	rm -f blog blog.db blog.db-wal blog.db-shm

vet: ## Run go vet
	go vet ./...
