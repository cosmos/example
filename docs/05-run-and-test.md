# Run, Test, and Configure

Now that you've [built a module from scratch](./03-build-a-module.md) and walked through the [full counter module](./04-counter-walkthrough.md), the next step is learning the workflow for running and validating a production-ready chain. This page shows how to start the chain locally, interact with it through the CLI, and use the main layers of testing before shipping changes.

## Single-node local chain

Use a single-node chain for the fastest local development loop. It gives you one validator with predictable state so you can quickly test queries and transactions.

### Start

```bash
make start
```

This builds the binary, initializes chain data, and starts a single validator node. It handles cleanup automatically — existing chain state is reset on each run.

The chain uses:
- Chain ID: `demo`
- Pre-funded accounts: `alice`, `bob`
- Default denomination: `stake`

### Stop

Press `Ctrl+C` in the terminal running `make start`.

### Reset chain state

```bash
make start
```

Re-running `make start` resets state automatically. There is no separate reset command.

## Localnet (multi-validator)

Use localnet when you want a setup that is closer to a real network. It runs multiple validators in Docker so you can test multi-node behavior locally.

For a multi-validator setup using Docker:

```bash
# Initialize localnet configuration
make localnet-init

# Start all validators
make localnet-start

# View logs
make localnet-logs

# Stop
make localnet-stop

# Clean all localnet data
make localnet-clean
```

## CLI reference

Once the chain is running, these are the core [CLI](https://docs.cosmos.network/sdk/next/learn/concepts/cli-grpc-rest#cli) commands you'll use to inspect state and submit transactions.

### Query commands

Use query commands to read module state without changing anything on-chain.

```bash
# Query the current counter value
exampled query counter count

# Query the module parameters
exampled query counter params

# Query with a specific node (if not using default localhost:26657)
exampled query counter count --node tcp://localhost:26657
```

### Transaction commands

Use transaction commands to submit state-changing messages to the chain.

```bash
# Add to the counter
exampled tx counter add 10 --from alice --chain-id demo --yes

# Add with a gas limit
exampled tx counter add 10 --from alice --chain-id demo --gas 200000 --yes

# Update module parameters (requires governance authority)
exampled tx counter update-params --from alice --chain-id demo --yes
```

### Useful flags

These flags are the ones you'll use most often while iterating locally.

| Flag | Description |
|---|---|
| `--from` | Key name or address to sign with |
| `--chain-id` | Chain ID (use `demo` for local) |
| `--yes` | Skip confirmation prompt |
| `--gas` | Gas limit for the transaction |
| `--node` | RPC endpoint (default: `tcp://localhost:26657`) |
| `--output json` | Output response as JSON |


## Node Configuration

When you run `make start`, the chain creates `~/.exampleapp/config/` automatically and initializes two config files inside it:

| File | What it controls |
|---|---|
| `app.toml` | SDK application settings: gas prices, pruning, API/gRPC servers, telemetry |
| `config.toml` | CometBFT settings: peer networking, consensus timeouts, mempool, RPC |

### app.toml

The most common settings to change during development:

| Setting | Default | Description |
|---|---|---|
| `minimum-gas-prices` | `"0stake"` | Minimum fee the node accepts before processing a transaction |
| `pruning` | `"default"` | How much historical state to keep (`default`, `nothing`, `everything`, `custom`) |
| `api.enable` | `true` | Enables the REST API on port 1317 |
| `grpc.enable` | `true` | Enables the gRPC server on port 9090 |

### config.toml

The settings most likely to change during development:

| Setting | Default | Description |
|---|---|---|
| `moniker` | `"test"` | Human-readable name for the node |
| `log_level` | `"info"` | Log verbosity (`debug`, `info`, `error`) |
| `consensus.timeout_commit` | `"5s"` | How long to wait after a block is committed before starting the next one |
| `p2p.seeds` | `""` | Seed nodes to connect to on a live network |
| `p2p.persistent_peers` | `""` | Peers to maintain permanent connections to |

## Unit tests

Start here when you want fast feedback on module logic without running a chain. These tests isolate the [keeper](https://docs.cosmos.network/sdk/next/learn/concepts/testing#keeper-unit-tests) and gRPC servers from the rest of the app.

The unit test logic lives in the counter keeper package on `main`: the shared suite setup is in [x/counter/keeper/keeper_test.go](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper_test.go), message-path tests are in [x/counter/keeper/msg_server_test.go](https://github.com/cosmos/example/blob/main/x/counter/keeper/msg_server_test.go), and query-path tests are in [x/counter/keeper/query_server_test.go](https://github.com/cosmos/example/blob/main/x/counter/keeper/query_server_test.go).

The keeper test suite covers the keeper, msg server, and query server in isolation using an in-memory store and a mock bank keeper. No running chain is required.

```bash
go test ./x/counter/...
```

To run with verbose output:

```bash
go test -v ./x/counter/...
```

To run a specific test:

```bash
go test -v -run TestKeeperTestSuite/TestAddCount ./x/counter/...
```

The test suite is structured around three files:

| File | Tests |
|---|---|
| `keeper/keeper_test.go` | Genesis, `GetCount`, `AddCount`, `SetParams` |
| `keeper/msg_server_test.go` | `MsgAdd`, event emission, `MsgUpdateParams` |
| `keeper/query_server_test.go` | `QueryCount`, `QueryParams` |

## E2E tests

Run [E2E tests](https://docs.cosmos.network/sdk/next/learn/concepts/testing#integration-tests) when you want to verify the full request path against a real node. They give you higher confidence than unit tests, but take longer to complete.

The E2E logic lives on `main` in [tests/counter_test.go](https://github.com/cosmos/example/blob/main/tests/counter_test.go), which starts an in-process network, builds signed transactions, and verifies query results. The shared network fixture it uses is defined in [tests/test_helpers.go](https://github.com/cosmos/example/blob/main/tests/test_helpers.go).

The E2E test suite starts a real in-process validator network and submits actual transactions against it. This tests the full stack: transaction encoding, message routing, keeper logic, and query responses.

```bash
go test -v -run TestE2ETestSuite ./tests/...
```

E2E tests take longer than unit tests because they spin up a real node. Run them before merging significant changes.

## Simulation tests

[Simulation tests](https://docs.cosmos.network/sdk/next/learn/concepts/testing#simulation-tests) stress the chain with randomized activity to catch edge cases that targeted tests can miss. In this repo, that simulation flow is built with `simsx`, the Cosmos SDK's higher-level simulation framework for defining random on-chain activity at the module level.

The top-level simulation test commands on `main` run through [sim_test.go](https://github.com/cosmos/example/blob/main/sim_test.go). The counter module's `simsx` registration lives in [x/counter/module.go](https://github.com/cosmos/example/blob/main/x/counter/module.go), the random `MsgAdd` generation lives in [x/counter/simulation/msg_factory.go](https://github.com/cosmos/example/blob/main/x/counter/simulation/msg_factory.go), and randomized counter genesis lives in [x/counter/simulation/genesis.go](https://github.com/cosmos/example/blob/main/x/counter/simulation/genesis.go).

In practice, `simsx` lets each module describe three things: how to generate random starting state, which operations can happen during simulation, and how often each operation should be chosen. For `x/counter`, that means generating a random initial counter value, registering `MsgAdd` as a simulation operation, and assigning it a weight so the simulator knows how frequently to try it relative to other module operations.

When you run a simulation target, the test harness repeatedly builds app instances, creates random accounts and balances, generates random transactions from the registered module operations, and executes them over many blocks. That makes `simsx` useful for catching issues that are hard to cover with hand-written tests, like state machine bugs, unexpected panics, invariant violations, and non-deterministic behavior across runs.

Simulation runs the chain with randomly generated transactions to detect non-determinism and invariant violations.

```bash
# Full simulation
make test-sim-full

# Determinism check
make test-sim-determinism

# All simulation tests
make test-sim
```

Simulation requires the `sims` build tag, which the Makefile targets handle automatically.

## Lint

Linting is the quickest way to catch style problems and common code-quality issues before CI or code review does.

The lint commands are defined in the repo [Makefile](https://github.com/cosmos/example/blob/main/Makefile), which installs `golangci-lint` and runs it across the full module tree.

```bash
make lint
```

This installs and runs `golangci-lint` across the repository. To auto-fix issues where possible:

```bash
make lint-fix
```

## Test summary

Use this table as a quick reference for choosing the right validation command for the kind of change you made.

| Command | What it validates |
|---|---|
| `go test ./x/counter/...` | Keeper, MsgServer, QueryServer in isolation |
| `go test -run TestE2ETestSuite ./tests/...` | Full transaction and query flow on a live node |
| `make test-sim-full` | Non-determinism and invariant violations |
| `make lint` | Code style and static analysis |
