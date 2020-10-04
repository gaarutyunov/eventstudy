BUILD = $(CURDIR)/build
LINT_FILE = $(CURDIR)/lint.toml
PROJECT_NAME = eventstudy

help: ## Show help dialog
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: install
install: ## Install dependencies
	go get -u github.com/mgechev/revive; \
 	go mod download

.PHONY: build
build: ## Build the project
	pkger; \
	go build -o $(BUILD)/$(PROJECT_NAME) $(CURDIR)/cmd/main.go

.PHONY: clean
clean: ## Clean project
	go clean; \
	rm -rf $(BUILD)

.PHONY: run
run: ## Run project locally
	$(BUILD)/$(PROJECT_NAME)

.PHONY: fmt
fmt: ## Format project
	go fmt $(CURDIR)/...

.PHONY: lint
lint: ## Lint project
	revive -config $(LINT_FILE) -formatter friendly $(CURDIR)/...

.PHONY: check
check: ## Format and lint project
check: fmt lint

.PHONY: setup
setup: ## Setup project
setup: clean install


