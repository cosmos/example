DOCKER := $(shell which docker)
PROTO_BUILDER_IMAGE := example-proto-builder

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=example \
	-X github.com/cosmos/cosmos-sdk/version.AppName=exampled \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

build:
	go build -ldflags '$(ldflags)' -o ./build/myapp ./exampled


install:
	go install -ldflags '$(ldflags)' ./exampled

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

# simsx runs every simulation across 38 built-in seeds. The SDK defaults of 500
# blocks x 200 operations per seed do not finish in any reasonable time, so the
# targets below use smaller values. Override to simulate more deeply, e.g.
#   make test-sim-full SIM_NUM_BLOCKS=500 SIM_BLOCK_SIZE=200 SIM_TIMEOUT=4h
SIM_NUM_BLOCKS ?= 50
SIM_BLOCK_SIZE ?= 100
SIM_TIMEOUT ?= 30m
SIM_FLAGS = -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE)

test-sim-full:
	@echo "--> Running full app simulation"
	go test -tags sims -run TestFullAppSimulation -v -timeout $(SIM_TIMEOUT) $(SIM_FLAGS)

test-sim-determinism:
	@echo "--> Running determinism simulation"
	go test -tags sims -run TestAppStateDeterminism -v -timeout $(SIM_TIMEOUT) $(SIM_FLAGS)

test-sim:
	@echo "--> Running all simulation tests"
	go test -tags sims -v -timeout $(SIM_TIMEOUT) $(SIM_FLAGS)

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