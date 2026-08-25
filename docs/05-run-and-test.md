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

## Localnet (multi-node)

Localnet runs four nodes in Docker to give you a setup closer to a real network than the single-node chain. `scripts/localnet/init.sh` creates a genesis transaction for `node0` only, so the network is **one validator plus three full nodes**, not four validators. The chain ID is `example-localnet`, and each node has a single key named `validator` rather than the `alice` and `bob` accounts used by `make start`.

Before you begin, note that this section needs Docker running, and that the following host ports must be free: `26656`, `26657`, `1317`, `9090` for `node0`, then `26666`, `26667`, `1318`, `9091` for `node1`, `26676`, `26677`, `1319`, `9092` for `node2`, and `26686`, `26687`, `1320`, `9093` for `node3`.

```bash
# Build the node image and initialize four node directories under build/localnet.
# Takes several minutes the first time, since it compiles the chain in Docker.
make localnet-init

# Start all four nodes
make localnet-start

# Follow the logs. This does not exit on its own; press Ctrl+C to stop following
make localnet-logs

# Stop
make localnet-stop

# Delete build/localnet immediately, without confirming
make localnet-clean
```

### Confirm the network is healthy

Each node exposes its own RPC port. Check that every node has found the other three and that they are advancing together:

```bash
for port in 26657 26667 26677 26687; do
  curl -s http://localhost:$port/status | grep -o '"latest_block_height":"[0-9]*"'
  curl -s http://localhost:$port/net_info | grep -o '"n_peers":"[0-9]*"'
done
```

Each node should report `"n_peers":"3"` and a block height that climbs on repeated calls.

### Send a transaction

The localnet uses a different chain ID and key name than `make start`, so the commands in the CLI reference below need adjusting. Run them inside a container:

```bash
docker exec node0 exampled tx counter add 7 \
  --from validator --chain-id example-localnet \
  --keyring-backend test --home /data/node0 --yes
```

Then confirm the state replicated by querying a different node:

```bash
docker exec node2 exampled query counter count --home /data/node2
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

```

### Updating module parameters

Counter params are governance-gated. `MsgUpdateParams` accepts only the gov module address as its
authority, so there is no direct CLI command for it: signing `update-params` with a user key such as
`alice` always fails with `ErrInvalidSigner`. Params change through a governance proposal instead.

Look up the gov module address for your chain, which is the only valid authority:

```bash
exampled query auth module-account gov
```

Write a `proposal.json` containing the message, using that address as `authority`. On the local `demo`
chain the value is `cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn`:

```json
{
  "messages": [
    {
      "@type": "/example.counter.MsgUpdateParams",
      "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
      "params": {
        "max_add_value": "50",
        "add_cost": [{"denom": "stake", "amount": "200"}]
      }
    }
  ],
  "metadata": "ipfs://CID",
  "deposit": "10000000stake",
  "title": "Update counter params",
  "summary": "Set max_add_value to 50 and add_cost to 200stake"
}
```

The `deposit` must meet the chain's `min_deposit`, which is `10000000stake` locally. Check it with
`exampled query gov params`. Then submit and vote:

```bash
exampled tx gov submit-proposal proposal.json --from alice --chain-id demo --yes
exampled tx gov vote 1 yes --from alice --chain-id demo --yes
```

Check progress with `exampled query gov proposals`.

<Note>
The local chain uses the default 48 hour `voting_period`, so a proposal submitted this way sits in
`PROPOSAL_STATUS_VOTING_PERIOD` for two days and the params do not change during a normal dev session.
</Note>

To watch a param change actually take effect locally, shorten the voting period. Editing
`genesis.json` before `make start` does not work, because `scripts/local_node.sh` deletes the whole
home directory on every run. Let `make start` create the chain first, then stop it and edit in place:

```bash
# 1. Let make start create ~/.exampleapp, then stop it with Ctrl+C
make start

# 2. Lower the voting period in the generated genesis
#    app_state.gov.params.voting_period, for example "20s"
vi ~/.exampleapp/config/genesis.json

# 3. Wipe block history so the edited genesis is re-read, keeping keys and config
exampled comet unsafe-reset-all

# 4. Start the node directly. Do not use make start again, it would delete your edit
exampled start
```

Submit and vote as above, wait out the shortened period, and the proposal reaches
`PROPOSAL_STATUS_PASSED` and `exampled query counter params` reflects the new values.

`exampled tx gov draft-proposal` can generate a skeleton, but it is an interactive terminal picker rather
than a scriptable command. Its top-level list offers only `text`, `community-pool-spend`,
`software-upgrade`, `cancel-software-upgrade`, and `other`, and choosing `other` opens a scroll-only list
of fully qualified message type URLs that typing does not filter. Writing the JSON by hand, as above, is
the more direct path.

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
| `minimum-gas-prices` | `"0stake"` | Minimum fee the node accepts before processing a transaction. Set by this chain in `exampled/cmd/commands.go`, not by the SDK, whose own default is empty |
| `pruning` | `"default"` | How much historical state to keep (`default`, `nothing`, `everything`, `custom`) |
| `api.enable` | `true` after `make start` | Enables the REST API on port 1317. The SDK default is `false`; `scripts/local_node.sh` turns it on for local development |
| `grpc.enable` | `true` | Enables the gRPC server on port 9090 |

### config.toml

The settings most likely to change during development:

| Setting | Default | Description |
|---|---|---|
| `moniker` | `"test"` | Human-readable name for the node |
| `log_level` | `"info"` | Log verbosity (`debug`, `info`, `error`) |
| `consensus.timeout_commit` | `"5s"` | How long to wait after a block is committed before starting the next one. The SDK raises CometBFT's own 1s default to 5s |
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

Each of these runs the simulation across 38 built-in seeds, so expect roughly ten minutes per target. The Makefile deliberately uses smaller values than the SDK defaults of 500 blocks and 200 operations per block, which across 38 seeds take hours. To simulate more deeply, override them:

```bash
make test-sim-full SIM_NUM_BLOCKS=500 SIM_BLOCK_SIZE=200 SIM_TIMEOUT=4h
```

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

| Command | What it validates | Typical runtime |
|---|---|---|
| `go test ./x/counter/...` | Keeper, MsgServer, QueryServer in isolation | seconds |
| `go test -run TestE2ETestSuite ./tests/...` | Full transaction and query flow on a live node | under a minute |
| `make test-sim-full` | Non-determinism and invariant violations | around ten minutes |
| `make lint` | Code style and static analysis | a few minutes |
