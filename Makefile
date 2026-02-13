build:
	go build cmd

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=latest

lint-install:
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)

lint:
	@echo "--> Running linter on all files"
	@$(MAKE) lint-install
	@golangci-lint run ./... --timeout=15m

lint-fix:
	@echo "--> Running linter"
	@$(MAKE) lint-install
	@golangci-lint run ./... --fix

.PHONY: lint lint-fix