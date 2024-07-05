DC = docker compose
DB_DSN = postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable
SERVER_PORT = 8080
GOPHERMARTTEST = gophermarttest \
	-test.v -test.run=^TestGophermart$$ \
	-gophermart-binary-path=cmd/gophermart/gophermart${EXE_POSTFIX} \
	-gophermart-host=localhost \
	-gophermart-port=$(SERVER_PORT) \
	-gophermart-database-uri="$(DB_DSN)" \
	-accrual-binary-path=cmd/accrual/$(ACCRUAL_BIN) \
	-accrual-host=localhost \
	-accrual-port=8097 \
	-accrual-database-uri="$(DB_DSN)"

ifeq ($(OS),Windows_NT)
    EXE_POSTFIX = .exe
    ACCRUAL_BIN = accrual_windows_amd64.exe
else
    EXE_POSTFIX =
    ACCRUAL_BIN = accrual_linux_amd64
endif

help: ## Show help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\%-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

up: ## Run docker containers
	$(DC) up -d

update: ## Update vendor, contracts, db-structure and everything else after switching to new code version
	go get ./...
	go mod vendor
update-packages: ## Update go modules versions
	go get -u ./...
	go mod tidy
	go mod vendor

lint: ## Run linter with settings from .golangci.yml
	golangci-lint run -v
lint-fix: ## Linter tries to fix issues automatically
	golangci-lint run -v --fix

.PHONY: test
test: ## Run local tests
	go test -v ./...
autotest: build ## Run autotest
	$(GOPHERMARTTEST)
cover:
	go test -cover ./...

build: ## Build server
	go build -C cmd/gophermart -o gophermart${EXE_POSTFIX}
run: ## Run server
	go run ./cmd/gophermart
