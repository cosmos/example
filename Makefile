DOCKER := $(shell which docker)
PROTO_BUILDER_IMAGE := example-proto-builder

build:
	go build -o ./build/myapp ./exampled


install:
	go install ./exampled

start: install
	./scripts/local_node.sh
	exampled start


proto-image-build:
	@docker build -t $(PROTO_BUILDER_IMAGE) -f proto/Dockerfile .

proto-gen:
	@echo "Generating Protobuf files"
	@$(DOCKER) run --rm -u 0 -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE) sh ./scripts/protocgen.sh

test:
	go test ./...

.PHONY: build install start test proto-image-build proto-gen

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

###############################################################################
###                                Simulation                               ###
###############################################################################

test-sim-full:
	@echo "--> Running full app simulation"
	go test -tags sims -run TestFullAppSimulation -v -timeout 30m

test-sim-determinism:
	@echo "--> Running determinism simulation"
	go test -tags sims -run TestAppStateDeterminism -v -timeout 30m

test-sim:
	@echo "--> Running all simulation tests"
	go test -tags sims -v -timeout 60m

.PHONY: test-sim-full test-sim-determinism test-sim

###############################################################################
###                              Docker / Localnet                          ###
###############################################################################

DOCKER_IMAGE := example-node

build-docker:
	@echo "--> Building Docker image $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .

localnet-init: build-docker
	@echo "--> Initializing localnet"
	@chmod +x scripts/localnet/init.sh
	@./scripts/localnet/init.sh

localnet-start:
	@echo "--> Starting localnet"
	docker compose up -d

localnet-stop:
	@echo "--> Stopping localnet"
	docker compose down

localnet-clean:
	@echo "--> Cleaning localnet data"
	rm -rf ./build/localnet

localnet-logs:
	docker compose logs -f

localnet: localnet-init localnet-start
	@echo "--> Localnet is running"

.PHONY: build-docker localnet-init localnet-start localnet-stop localnet-clean localnet-logs localnet