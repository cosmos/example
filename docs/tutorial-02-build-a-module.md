# Build a Module from Scratch

This tutorial walks through building a minimal counter module from scratch so you can learn the core Cosmos SDK module pattern as quickly as possible. It uses the same overall structure as `x/counter`, the fuller example module in this repository.


By the end, you'll have a working counter module wired into a running chain.

## Before You Begin

<!-- todo: implement ideass in ideas/ideas.md -->

This tutorial uses the `tutorial/start` branch, which has the counter module stripped out and `app.go` wiring removed. The directory structure is in place — you fill it in.

```bash
git clone https://github.com/cosmos/example
cd example
git checkout tutorial/start
mkdir -p x/counter/keeper x/counter/types proto/example/counter/v1
```

You should see empty placeholder directories at `x/counter/` and `proto/example/counter/v1/`.

> **Branch model:** On the `main` branch, `x/counter` is the full production module. On `tutorial/start`, that module and its app.go wiring are removed so you can rebuild a minimal version from scratch under the same path (`x/counter`). Without this context, it may look like you are editing the production module — you are not.


## The Module Loop

Every Cosmos SDK module follows the same pattern:

```text
proto files → code generation → keeper → msg server → query server → module.go → app wiring
```

- **Proto files** define the module's messages, queries, and state types
- **Code generation** (`make proto-gen`) produces Go interfaces and types from those definitions
- **Keeper** owns and manages the module's state
- **MsgServer** handles incoming transactions and delegates to the keeper
- **QueryServer** handles read-only queries against the keeper
- **module.go** wires everything together and registers it with the application

Here is how each proto definition maps to generated code and then to your implementation:

| Proto | Generated (types/) | Your implementation |
| --- | --- | --- |
| `service Msg { rpc Add }` | `MsgServer` interface | `keeper/msg_server.go` |
| `service Query { rpc Count }` | `QueryServer` interface | `keeper/query_server.go` |
| `message MsgAddRequest` | `MsgAddRequest` struct | used as input in `msg_server.go` |
| `message GenesisState` | `GenesisState` struct | used in `InitGenesis` / `ExportGenesis` |

Steps 3–8 implement the right-hand column. Steps 1–2 produce the middle column.


## Directory Structure

```text
x/counter/
├── keeper/
│   ├── keeper.go         # Keeper struct and state methods
│   ├── msg_server.go     # MsgServer implementation
│   └── query_server.go   # QueryServer implementation
├── types/
│   ├── keys.go           # Module name and store key constants
│   ├── codec.go          # Interface registration
│   └── *.pb.go           # Generated from proto — do not edit
├── module.go             # AppModule wiring
└── autocli.go            # CLI command definitions

proto/example/counter/v1/
├── tx.proto
├── query.proto
└── genesis.proto
```

## Step 1: Proto files

Proto files are the source of truth for the module's public API. You define messages and services here. In this tutorial, the counter module stores one number, `Add` increases it by the amount the user submits, and the query returns the current value. After, `make proto-gen` produces the Go interfaces you then implement.

First, create the three proto files:

```bash
touch proto/example/counter/v1/tx.proto \
  proto/example/counter/v1/query.proto \
  proto/example/counter/v1/genesis.proto
```

Then add the following contents to each file.

### tx.proto

This is the first module file you define. It declares the transaction message shape for `Add`: what the user sends to increment the counter, and what the module returns after handling it.

```proto
syntax = "proto3";

// Matches the module's protobuf namespace.
package example.counter;

// Provides Cosmos SDK message annotations like signer and service markers.
import "cosmos/msg/v1/msg.proto";

// Generated Go types are written into x/counter/types.
option go_package = "github.com/cosmos/example/x/counter/types";

service Msg {
  // Marks this as a transaction service, not a normal gRPC service.
  option (cosmos.msg.v1.service) = true;
  // Add is the one transaction this minimal module supports.
  rpc Add(MsgAddRequest) returns (MsgAddResponse);
}

message MsgAddRequest {
  // The sender signs this message.
  option (cosmos.msg.v1.signer) = "sender";
  string sender = 1;
  uint64 add    = 2;
}

message MsgAddResponse {
  // Return the new counter value after the add succeeds.
  uint64 updated_count = 1;
}
```

### query.proto

This file defines the read-only gRPC query service and the response type for fetching the current count.

```proto
syntax = "proto3";

// Matches the module's protobuf namespace.
package example.counter;

// Enables the REST gateway route annotation below.
import "google/api/annotations.proto";

// Generated Go types are written into x/counter/types.
option go_package = "github.com/cosmos/example/x/counter/types";

service Query {
  rpc Count(QueryCountRequest) returns (QueryCountResponse) {
    // Exposes this query over the HTTP API as well as gRPC.
    option (google.api.http).get = "/example/counter/v1/count";
  }
}

// Empty because this query only needs the module's current state.
message QueryCountRequest  {}

message QueryCountResponse {
  // The current counter value.
  uint64 count = 1;
}
```

### genesis.proto

This file defines the data the module stores in genesis so the counter can be initialized when the chain starts.

```proto
syntax = "proto3";

// Matches the module's protobuf namespace.
package example.counter;

// Generated Go types are written into x/counter/types.
option go_package = "github.com/cosmos/example/x/counter/types";

message GenesisState {
  // The counter value to load when the chain initializes.
  uint64 count = 1;
}
```

## Step 2: Generate Code

Make sure Docker is running. The first time you run proto-gen you need to build the builder image:

```bash
make proto-image-build
make proto-gen
```

This compiles the proto files using [buf](https://buf.build) inside Docker. The generated files will appear in `x/counter/types/`:

```text
x/counter/types/
├── tx.pb.go         # MsgAddRequest, MsgAddResponse, MsgServer interface
├── query.pb.go      # QueryCountRequest, QueryCountResponse, QueryServer interface
├── query.pb.gw.go   # REST gateway registration
└── genesis.pb.go    # GenesisState
```

> **Do not edit generated files.** Changes to public types belong in the proto files. Re-run `make proto-gen` after any proto change.

The most important generated output is the `MsgServer` and `QueryServer` interfaces. In Steps 5 and 6, you'll implement them in `keeper/msg_server.go` and `keeper/query_server.go`.


## Step 3: Types

In this step, you define the small supporting files in `x/counter/types` that the rest of the module depends on before you write any keeper logic. They live here because they describe shared module types and identifiers, not state-management code.

First, create the two files for this section:

```bash
touch x/counter/types/keys.go \
  x/counter/types/codec.go
```

Then add the following contents to each file.

### keys.go

This file defines the module's basic identifiers: the module name used throughout the SDK, and the store key used to claim the module's KV store namespace.

```go
// x/counter/types/keys.go
package types

const (
    // ModuleName is the name the SDK uses to refer to this module.
    ModuleName = "counter"
    // StoreKey is the key for this module's KV store.
    StoreKey   = ModuleName
)
```

`ModuleName` identifies the module throughout the SDK (routing, events, governance). `StoreKey` is the key used to claim the module's isolated namespace in the chain's KV store (set equal to `ModuleName` by convention).

### Interface Registration

This file registers your generated message types with the SDK interface registry so the application can decode and route your module's transactions correctly.

```go
// x/counter/types/codec.go
package types

import (
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk        "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
    // Register MsgAddRequest as an sdk.Msg so the app can decode it from transactions.
    registry.RegisterImplementations((*sdk.Msg)(nil),
        &MsgAddRequest{},
    )
    // Register the generated Msg service description for routing.
    msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
```

`_Msg_serviceDesc` is generated by `make proto-gen` — it describes the `Msg` gRPC service defined in `tx.proto`.


## Step 4: Keeper

In this step, you create the keeper, which is the part of the module that owns the counter state and provides the methods the rest of the module will call.

Create the keeper file:

```bash
touch x/counter/keeper/keeper.go
```

Then add the following contents.

This file defines the keeper struct, sets up the counter's storage item, and implements the core state methods for reading, updating, and loading the counter at genesis.

```go
// x/counter/keeper/keeper.go
package keeper

import (
    "context"
    "errors"

    "cosmossdk.io/collections"
    "cosmossdk.io/core/store"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/example/x/counter/types"
)

type Keeper struct {
    Schema  collections.Schema
    counter collections.Item[uint64]
}

func NewKeeper(storeService store.KVStoreService, cdc codec.Codec) *Keeper {
    sb := collections.NewSchemaBuilder(storeService)
    k := Keeper{
        // Store the counter under prefix 0 in this module's KV store.
        counter: collections.NewItem(sb, collections.NewPrefix(0), "counter", collections.Uint64Value),
    }
    schema, err := sb.Build()
    if err != nil {
        panic(err)
    }
    k.Schema = schema
    return &k
}

func (k *Keeper) GetCount(ctx context.Context) (uint64, error) {
    count, err := k.counter.Get(ctx)
    // Treat missing state as zero so a fresh chain starts cleanly.
    if err != nil && !errors.Is(err, collections.ErrNotFound) {
        return 0, err
    }
    return count, nil
}

func (k *Keeper) AddCount(ctx context.Context, amount uint64) (uint64, error) {
    count, err := k.GetCount(ctx)
    if err != nil {
        return 0, err
    }
    // Increment the current count and write it back to state.
    newCount := count + amount
    return newCount, k.counter.Set(ctx, newCount)
}

func (k *Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
    return k.counter.Set(ctx, gs.Count)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
    count, err := k.GetCount(ctx)
    if err != nil {
        return nil, err
    }
    return &types.GenesisState{Count: count}, nil
}
```

`collections.Item[uint64]` is a typed KV store entry; the `collections` package handles encoding and namespacing. `GetCount` treats `ErrNotFound` as zero so the counter starts at zero without explicit initialization.

> **State layout**
>
> - `StoreKey` (`"counter"`) is the module's isolated namespace within the chain's global KV store. No other module can read or write this namespace.
> - `collections.NewPrefix(0)` is a single-byte prefix that identifies the `counter` item within the module's namespace. A module with multiple items would use `NewPrefix(0)`, `NewPrefix(1)`, etc. to keep them separate.
> - `ErrNotFound` treated as zero means the keeper never needs to explicitly set an initial value — the first `GetCount` call on a fresh chain returns `0` by convention.


## Step 5: MsgServer

In this step, you implement the transaction handler for the generated `MsgServer` interface. This is the code path that runs when a user submits `tx counter add`.

Create the message server file:

```bash
touch x/counter/keeper/msg_server.go
```

Then add the following contents.

This file implements the generated `MsgServer` interface and forwards the `Add` transaction to the keeper's `AddCount` method.

```go
// x/counter/keeper/msg_server.go
package keeper

import (
    "context"

    "github.com/cosmos/example/x/counter/types"
)

type msgServer struct {
    *Keeper
}

func NewMsgServerImpl(k *Keeper) types.MsgServer {
    return &msgServer{k}
}

func (m msgServer) Add(ctx context.Context, req *types.MsgAddRequest) (*types.MsgAddResponse, error) {
    // Delegate the state update to the keeper.
    newCount, err := m.AddCount(ctx, req.GetAdd())
    if err != nil {
        return nil, err
    }
    // Return the updated count back to the caller.
    return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}
```

`msgServer` embeds `*Keeper` and delegates directly to `AddCount`. The handler itself contains no business logic.


## Step 6: QueryServer

In this step, you implement the read-only query handler for the generated `QueryServer` interface. This is the code path that runs when someone queries the current counter value.

Create the query server file:

```bash
touch x/counter/keeper/query_server.go
```

Then add the following contents.

This file implements the generated `QueryServer` interface and returns the current counter value from the keeper.

```go
// x/counter/keeper/query_server.go
package keeper

import (
    "context"

    "github.com/cosmos/example/x/counter/types"
)

type queryServer struct {
    *Keeper
}

func NewQueryServer(k *Keeper) types.QueryServer {
    return &queryServer{k}
}

func (q queryServer) Count(ctx context.Context, _ *types.QueryCountRequest) (*types.QueryCountResponse, error) {
    // Read the current count from state and return it in the query response.
    count, err := q.GetCount(ctx)
    if err != nil {
        return nil, err
    }
    return &types.QueryCountResponse{Count: count}, nil
}
```


## Step 7: module.go

In this step, you connect your keeper and generated services to the Cosmos SDK module framework so the application knows how to initialize the module, expose its query routes, and register its transaction handlers.

Create the module file:

```bash
touch x/counter/module.go
```

Then add the following contents.

This file defines the app module types and wires your keeper into genesis handling, service registration, and gRPC gateway registration.

```go
// x/counter/module.go
package counter

import (
    "context"
    "encoding/json"

    "cosmossdk.io/core/appmodule"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"

    "github.com/cosmos/example/x/counter/keeper"
    countertypes "github.com/cosmos/example/x/counter/types"
)

var (
    // Compile-time checks that AppModule implements the required module interfaces.
    _ appmodule.AppModule        = AppModule{}
    _ module.HasConsensusVersion = AppModule{}
    _ module.HasGenesis          = AppModule{}
    _ module.HasServices         = AppModule{}
)

type AppModuleBasic struct {
    cdc codec.Codec
}

func (a AppModuleBasic) Name() string { return countertypes.ModuleName }

func (a AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (a AppModuleBasic) RegisterInterfaces(registry codecTypes.InterfaceRegistry) {
    countertypes.RegisterInterfaces(registry)
}

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    // Start the module with a zero counter by default.
    return cdc.MustMarshalJSON(&countertypes.GenesisState{Count: 0})
}

func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
    gs := countertypes.GenesisState{}
    return cdc.UnmarshalJSON(bz, &gs)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
    // Expose the Query service through the HTTP gateway.
    if err := countertypes.RegisterQueryHandlerClient(context.Background(), mux, countertypes.NewQueryClient(clientCtx)); err != nil {
        panic(err)
    }
}

type AppModule struct {
    AppModuleBasic
    keeper *keeper.Keeper
}

func NewAppModule(cdc codec.Codec, k *keeper.Keeper) AppModule {
    return AppModule{AppModuleBasic: AppModuleBasic{cdc: cdc}, keeper: k}
}

func (a AppModule) IsOnePerModuleType() {}
func (a AppModule) IsAppModule()        {}

func (a AppModule) ConsensusVersion() uint64 { return 1 }

func (a AppModule) RegisterServices(cfg module.Configurator) {
    // Connect the generated service interfaces to your keeper-backed implementations.
    countertypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
    countertypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(a.keeper))
}

func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
    gs := &countertypes.GenesisState{}
    cdc.MustUnmarshalJSON(bz, gs)
    // Load the initial counter value into state at chain start.
    if err := a.keeper.InitGenesis(ctx, gs); err != nil {
        panic(err)
    }
}

func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    gs, err := a.keeper.ExportGenesis(ctx)
    if err != nil {
        panic(err)
    }
    // Write the current counter value back out for exports.
    return cdc.MustMarshalJSON(gs)
}
```

The `var _ interface = Struct{}` block at the top is a Go compile-time check — if the struct is missing any required method, the build fails immediately.

`RegisterServices` is the most important method. It connects the generated server interfaces to your implementations, making them reachable from the SDK's message and query routers.


## Step 8: AutoCLI

In this step, you define the CLI metadata for your module. AutoCLI reads this configuration together with your proto services and generates the `exampled query counter` and `exampled tx counter` commands automatically.

Create the AutoCLI file:

```bash
touch x/counter/autocli.go
```

Then add the following contents.

This file tells AutoCLI how to expose the `Count` query and `Add` transaction as simple command-line commands.

```go
// x/counter/autocli.go
package counter

import (
    autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
    return &autocliv1.ModuleOptions{
        Query: &autocliv1.ServiceCommandDescriptor{
            Service: "example.counter.Query",
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                // exampled query counter count
                {RpcMethod: "Count", Use: "count", Short: "Query the current counter value"},
            },
        },
        Tx: &autocliv1.ServiceCommandDescriptor{
            Service: "example.counter.Msg",
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                // exampled tx counter add 5 --from alice
                {RpcMethod: "Add", Use: "add [amount]", Short: "Add to the counter",
                    PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "add"}}},
            },
        },
    }
}
```

`PositionalArgs` maps the first CLI argument to the `add` field in `MsgAddRequest`, so `add 5` works instead of `add --add 5`.


## Step 9: Wire into app.go

In this step, you wire your new module into the application so the chain creates its store, constructs its keeper, and includes it in module startup and genesis handling.

Open `app.go`. For each change below, find the matching block and add the highlighted lines.

### Imports

Find the import block near the bottom of the Cosmos SDK imports and add these three lines:

```go
// tutorial-02: add counter imports here
// After: "github.com/cosmos/cosmos-sdk/x/tx/signing"
	counter       "github.com/cosmos/example/x/counter"
	counterkeeper "github.com/cosmos/example/x/counter/keeper"
	countertypes  "github.com/cosmos/example/x/counter/types"
```

### Keeper Field

Find the keeper fields on `type ExampleApp struct` and add `CounterKeeper`:

```go
// tutorial-02: add counter keeper field here
// After: ConsensusParamsKeeper consensusparamkeeper.Keeper
CounterKeeper         *counterkeeper.Keeper
```

### Store Key

Find `keys := storetypes.NewKVStoreKeys(` and add the counter store key:

```go
// tutorial-02: add counter store key here
// After: consensusparamtypes.StoreKey,
countertypes.StoreKey,
```

### Keeper Instantiation

Find the code just after `app.GovKeeper = ...` and add the counter keeper construction:

```go
// tutorial-02: create the counter keeper here
// After: app.GovKeeper = *govKeeper.SetHooks(...)
app.CounterKeeper = counterkeeper.NewKeeper(
	runtime.NewKVStoreService(keys[countertypes.StoreKey]),
	appCodec,
)
```

### Module Manager

Find `app.ModuleManager = module.NewManager(` and add the counter module to the list:

```go
// tutorial-02: register the counter module here
// After: vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
counter.NewAppModule(appCodec, app.CounterKeeper),
```

### Genesis and Export Order

For the minimal module, the important ordering is genesis import and export. Add `countertypes.ModuleName` to both `genesisModuleOrder` and `exportModuleOrder`. This module has no special ordering requirements, so placing it near the end of each list is fine. The production module adds module-account wiring and begin/end block hooks later in Tutorial 03.

```go
// tutorial-02: add this in genesisModuleOrder and exportModuleOrder
countertypes.ModuleName,
```


## Step 10: Build

Run the following to compile the app and make sure the new module wiring is valid before you try to run the chain.

```bash
go build ./...
```

Fix any compilation errors before continuing.


## Step 11: Smoke Test

Now you'll run the app locally and use one transaction plus one query to confirm the module works end-to-end.

### Start the chain

First, install the binary and start the demo chain.

```bash
make install
make start
```

This builds and installs `exampled` and then runs [`scripts/local_node.sh`](../../scripts/local_node.sh), which:
- resets the local home directory
- initializes genesis
- creates the `alice` and `bob` test accounts
- funds those accounts
- creates a validator transaction
- starts the chain

You'll see the chain running and it should start producing blocks. 

### Submit a transaction

Open a second terminal and submit a transaction that adds `4` to the counter:

```bash
exampled tx counter add 4 --from alice --chain-id demo --yes
```

If the transaction succeeds, the response should include `code: 0`, which means the chain accepted and executed the transaction without an application error:

```
code: 0
```

### Query the chain

Query the counter to confirm the stored value changed. This uses the `exampled query counter count` command that AutoCLI generated from the `Query` service you defined earlier:

```bash
exampled query counter count
```

You should see the following output:

```text
count: "4"
```

Congratulations, You've just created a Cosmos module from scratch and wired it into a real chain!


## Summary

| File | Role |
| --- | --- |
| `tx.proto` / `query.proto` / `genesis.proto` | Declare the module's public API |
| `types/keys.go` | Module name and store key |
| `types/codec.go` | Interface registration |
| `keeper/keeper.go` | State ownership and access |
| `keeper/msg_server.go` | Transaction handling |
| `keeper/query_server.go` | Query handling |
| `module.go` | Framework registration |
| `autocli.go` | CLI command definitions |
| `app.go` | Chain integration |

The production `x/counter` on the `main` branch follows this exact same structure. In the next section you'll see what it adds on top.


Next: [Production Counter Walkthrough →](./tutorial-03-counter-walkthrough.md)
