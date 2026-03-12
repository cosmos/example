# Example Cosmos SDK Application

A minimal Cosmos SDK example application demonstrating best practices for building blockchain modules. This repository serves as a learning resource and starting point for new Cosmos SDK developers.

## Features

- **Custom Counter Module** - A simple module demonstrating the full module development lifecycle
- **Complete Test Suite** - Unit tests, integration tests, and simulation tests
- **AutoCLI Integration** - Automatic CLI command generation
- **Local Development** - Scripts for single-node and multi-node local networks
- **Docker Support** - Containerized builds and local network deployment

## Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for protobuf generation and localnet)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/cosmos/example.git
cd example

# Build the application
make build

# Install the binary
make install

# Initialize and start a local node
make start
```

## Project Structure

```
.
├── app.go                 # Application wiring and module registration
├── exampled/              # CLI entrypoint
│   ├── main.go
│   └── cmd/
│       ├── commands.go    # CLI commands wiring
│       └── root.go        # Root command setup
├── x/counter/             # Custom counter module
│   ├── autocli.go         # AutoCLI configuration
│   ├── module.go          # Module definition and interfaces
│   ├── keeper/            # State management
│   │   ├── keeper.go      # Keeper implementation
│   │   ├── msg_server.go  # Transaction handlers
│   │   └── query_server.go# Query handlers
│   ├── types/             # Proto-generated types
│   │   └── codec.go       # Type registration
│   └── simulation/        # Simulation testing
│       ├── genesis.go     # Randomized genesis
│       └── msg_factory.go # Message factories
├── proto/                 # Protobuf definitions
│   └── example/counter/v1/
│       ├── tx.proto       # Transaction messages
│       ├── query.proto    # Query definitions
│       └── genesis.proto  # Genesis state
├── tests/                 # Integration tests
├── scripts/               # Development scripts
└── Makefile               # Build and development commands
```

## The Counter Module

The counter module is a minimal example that demonstrates core Cosmos SDK patterns:

### State

A single `uint64` counter value stored using the `collections` library.

### Messages

| Message | Description |
|---------|-------------|
| `MsgAdd` | Increments the counter by a specified amount |

### Queries

| Query | Description |
|-------|-------------|
| `Count` | Returns the current counter value |

### Events

The module emits events when the counter is updated:

```json
{
  "type": "counter_added",
  "attributes": [
    {"key": "previous_count", "value": "0"},
    {"key": "added", "value": "5"},
    {"key": "new_count", "value": "5"}
  ]
}
```

## CLI Usage

After starting a local node, interact with the counter module:

```bash
# Query the current count
exampled query counter count

# Add to the counter (alice is a pre-funded account from the local_node.sh script in /scripts)
exampled tx counter add 5 --from alice

# Check transaction result
exampled query tx <txhash>
```

## Development

### Building

```bash
# Build binary to ./build/myapp
make build

# Install to $GOPATH/bin
make install
```

### Running a Local Node

```bash
# Single node (installs, initializes, and starts)
make start

# Or manually:
./scripts/local_node.sh  # Initialize
exampled start           # Start the node
```

### Running a Local Network (4 nodes)

```bash
# Initialize and start 4-node network
make localnet

# View logs
make localnet-logs

# Stop network
make localnet-stop

# Clean up
make localnet-clean
```

### Generating Protobuf Files

```bash
# Build the proto builder image (first time only)
make proto-image-build

# Generate Go code from proto files
make proto-gen
```

## Testing

### Unit Tests

```bash
go test ./x/counter/keeper/...
```

### Integration Tests

Test the module within a full application context:

```bash
go test ./tests/...
```

### Simulation Tests

Fuzz testing with randomized inputs and state:

```bash
# Run full simulation (all seeds)
make test-sim-full

# Run determinism test
make test-sim-determinism

# Run single seed (faster)
go test -tags sims -run "TestFullAppSimulation/seed:_1$" -v
```

### All Tests

```bash
go test ./...
```

### Linting

```bash
make lint      # Check for issues
make lint-fix  # Auto-fix issues
```

## Module Development Guide

### Adding a New Message

1. **Define the proto message** in `proto/example/counter/v1/tx.proto`:
   ```protobuf
   message MsgNewAction {
     option (cosmos.msg.v1.signer) = "sender";
     string sender = 1;
     // your fields
   }
   ```

2. **Regenerate proto files**:
   ```bash
   make proto-gen
   ```

3. **Implement the handler** in `x/counter/keeper/msg_server.go`:
   ```go
   func (m msgServer) NewAction(ctx context.Context, msg *types.MsgNewAction) (*types.MsgNewActionResponse, error) {
       // implementation
   }
   ```

4. **Add AutoCLI config** in `x/counter/autocli.go`

5. **Write tests** in `x/counter/keeper/msg_server_test.go`

6. **Add simulation** in `x/counter/simulation/msg_factory.go`

### Adding a New Query

1. **Define in proto** (`query.proto`)
2. **Regenerate**: `make proto-gen`
3. **Implement** in `keeper/query_server.go`
4. **Add AutoCLI config**
5. **Write tests**

## Included Modules

| Module | Description |
|--------|-------------|
| `auth` | Account management and authentication |
| `bank` | Token transfers and balances |
| `staking` | Proof-of-stake validator management |
| `distribution` | Reward distribution |
| `gov` | On-chain governance |
| `slashing` | Validator punishment |
| `consensus` | Consensus parameters |
| `vesting` | Vesting accounts |
| `counter` | Example custom module |

## Resources

- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Cosmos SDK GitHub](https://github.com/cosmos/cosmos-sdk)
- [CometBFT Documentation](https://docs.cometbft.com/)
- [Protobuf Guide](https://developers.google.com/protocol-buffers)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.
