# Running and Testing

This page covers how to run the chain locally, use the CLI, and execute the full test suite.

---

## Single-node local chain

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

---

## Localnet (multi-validator)

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

---

## CLI reference

### Query commands

```bash
# Query the current counter value
exampled query counter count

# Query the module parameters
exampled query counter params

# Query with a specific node (if not using default localhost:26657)
exampled query counter count --node tcp://localhost:26657
```

### Transaction commands

```bash
# Add to the counter
exampled tx counter add 10 --from alice --chain-id demo --yes

# Add with a gas limit
exampled tx counter add 10 --from alice --chain-id demo --gas 200000 --yes

# Update module parameters (requires governance authority)
exampled tx counter update-params --from alice --chain-id demo --yes
```

### Useful flags

| Flag | Description |
|---|---|
| `--from` | Key name or address to sign with |
| `--chain-id` | Chain ID (use `demo` for local) |
| `--yes` | Skip confirmation prompt |
| `--gas` | Gas limit for the transaction |
| `--node` | RPC endpoint (default: `tcp://localhost:26657`) |
| `--output json` | Output response as JSON |

---

## Unit tests

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

---

## E2E tests

The E2E test suite starts a real in-process validator network and submits actual transactions against it. This tests the full stack: transaction encoding, message routing, keeper logic, and query responses.

```bash
go test -v -run TestE2ETestSuite ./tests/...
```

E2E tests take longer than unit tests because they spin up a real node. Run them before merging significant changes.

---

## Simulation tests

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

---

## Lint

```bash
make lint
```

This installs and runs `golangci-lint` across the repository. To auto-fix issues where possible:

```bash
make lint-fix
```

---

## Test summary

| Command | What it validates |
|---|---|
| `go test ./x/counter/...` | Keeper, MsgServer, QueryServer in isolation |
| `go test -run TestE2ETestSuite ./tests/...` | Full transaction and query flow on a live node |
| `make test-sim-full` | Non-determinism and invariant violations |
| `make lint` | Code style and static analysis |
