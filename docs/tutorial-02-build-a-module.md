# Build a Module from Scratch

This tutorial walks through building a minimal counter module from the ground up. The goal is to understand the core SDK module pattern — the same pattern used by `x/counter`, the production module in this repository.

**Scope:** 1 state item, 1 message, 1 query. No parameters, no fees, no telemetry.

By the end you will have a working counter module wired into a running chain.

---

## Before You Begin

This tutorial uses the `tutorial/start` branch, which has the counter module stripped out and app.go wiring removed. The directory structure is in place — you fill it in.

```bash
git clone https://github.com/cosmos/example
cd example
git checkout tutorial/start
```

You should see empty placeholder directories at `x/counter/` and `proto/example/counter/v1/`.

> **Branch model:** On the `main` branch, `x/counter` is the full production module. On `tutorial/start`, that module and its app.go wiring are removed so you can rebuild a minimal version from scratch under the same path (`x/counter`). Without this context, it may look like you are editing the production module — you are not.

---

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

---

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

---

## Step 1 — Proto files

Proto files are the source of truth for the module's public API. You define messages and services here; `make proto-gen` produces the Go interfaces you then implement.

### tx.proto

```proto
syntax = "proto3";
package example.counter;

import "cosmos/msg/v1/msg.proto";

option go_package = "github.com/cosmos/example/x/counter/types";

service Msg {
  rpc Add(MsgAddRequest) returns (MsgAddResponse);
}

message MsgAddRequest {
  option (cosmos.msg.v1.signer) = "sender";
  string sender = 1;
  uint64 add    = 2;
}

message MsgAddResponse {
  uint64 updated_count = 1;
}
```

### query.proto

```proto
syntax = "proto3";
package example.counter;

import "google/api/annotations.proto";

option go_package = "github.com/cosmos/example/x/counter/types";

service Query {
  rpc Count(QueryCountRequest) returns (QueryCountResponse) {
    option (google.api.http).get = "/example/counter/v1/count";
  }
}

message QueryCountRequest  {}

message QueryCountResponse {
  uint64 count = 1;
}
```

### genesis.proto

```proto
syntax = "proto3";
package example.counter;

option go_package = "github.com/cosmos/example/x/counter/types";

message GenesisState {
  uint64 count = 1;
}
```

---

## Step 2 — Generate Code

Make sure Docker is running. The first time you run proto-gen you need to build the builder image:

```bash
make proto-image-build
make proto-gen
```

This compiles the proto files using [buf](https://buf.build) inside Docker. Generated files appear in `x/counter/types/`:

```text
x/counter/types/
├── tx.pb.go         # MsgAddRequest, MsgAddResponse, MsgServer interface
├── query.pb.go      # QueryCountRequest, QueryCountResponse, QueryServer interface
├── query.pb.gw.go   # REST gateway registration
└── genesis.pb.go    # GenesisState
```

> **Do not edit generated files.** Changes to public types belong in the proto files. Re-run `make proto-gen` after any proto change.

The key output is two interfaces you must implement in Steps 5 and 6:

```go
type MsgServer interface {
    Add(context.Context, *MsgAddRequest) (*MsgAddResponse, error)
}

type QueryServer interface {
    Count(context.Context, *QueryCountRequest) (*QueryCountResponse, error)
}
```

---

## Step 3 — Types

### keys.go

```go
// x/counter/types/keys.go
package types

const (
    ModuleName = "counter"
    StoreKey   = ModuleName
)
```

`ModuleName` identifies the module throughout the SDK (routing, events, governance). `StoreKey` is the key used to claim the module's isolated namespace in the chain's KV store — set equal to `ModuleName` by convention.

### Interface Registration

`module.go` calls `RegisterInterfaces` to register the module's message types with the SDK's interface registry. Without this, the chain cannot deserialize transactions containing your messages.

```go
// x/counter/types/codec.go
package types

import (
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk        "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
    registry.RegisterImplementations((*sdk.Msg)(nil),
        &MsgAddRequest{},
    )
    msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
```

`_Msg_serviceDesc` is generated by `make proto-gen` — it describes the `Msg` gRPC service defined in `tx.proto`.

---

## Step 4 — Keeper

The keeper owns the module's state. No other code touches the store directly.

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

`collections.Item[uint64]` is a typed KV store entry — the `collections` package handles encoding and namespacing. `GetCount` treats `ErrNotFound` as zero so the counter starts at zero without explicit initialization.

> **State layout**
>
> - `StoreKey` (`"counter"`) is the module's isolated namespace within the chain's global KV store. No other module can read or write this namespace.
> - `collections.NewPrefix(0)` is a single-byte prefix that identifies the `counter` item within the module's namespace. A module with multiple items would use `NewPrefix(0)`, `NewPrefix(1)`, etc. to keep them separate.
> - `ErrNotFound` treated as zero means the keeper never needs to explicitly set an initial value — the first `GetCount` call on a fresh chain returns `0` by convention.

---

## Step 5 — MsgServer

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
    newCount, err := m.AddCount(ctx, req.GetAdd())
    if err != nil {
        return nil, err
    }
    return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}
```

`msgServer` embeds `*Keeper` and delegates directly to `AddCount`. The handler itself contains no business logic.

---

## Step 6 — QueryServer

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
    count, err := q.GetCount(ctx)
    if err != nil {
        return nil, err
    }
    return &types.QueryCountResponse{Count: count}, nil
}
```

---

## Step 7 — module.go

`module.go` registers the module with the application framework. It implements a set of interfaces that tell the SDK how to initialize the module, handle genesis, and register its services.

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
    return cdc.MustMarshalJSON(&countertypes.GenesisState{Count: 0})
}

func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
    gs := countertypes.GenesisState{}
    return cdc.UnmarshalJSON(bz, &gs)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
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
    countertypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
    countertypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(a.keeper))
}

func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
    gs := &countertypes.GenesisState{}
    cdc.MustUnmarshalJSON(bz, gs)
    if err := a.keeper.InitGenesis(ctx, gs); err != nil {
        panic(err)
    }
}

func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    gs, err := a.keeper.ExportGenesis(ctx)
    if err != nil {
        panic(err)
    }
    return cdc.MustMarshalJSON(gs)
}
```

The `var _ interface = Struct{}` block at the top is a Go compile-time check — if the struct is missing any required method, the build fails immediately.

`RegisterServices` is the most important method. It connects the generated server interfaces to your implementations, making them reachable from the SDK's message and query routers.

---

## Step 8 — AutoCLI

`autocli.go` defines the CLI commands for your module. AutoCLI reads the proto service definitions and generates `exampled query counter` and `exampled tx counter` subcommands automatically.

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
                {RpcMethod: "Count", Use: "count", Short: "Query the current counter value"},
            },
        },
        Tx: &autocliv1.ServiceCommandDescriptor{
            Service: "example.counter.Msg",
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                {RpcMethod: "Add", Use: "add [amount]", Short: "Add to the counter",
                    PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "add"}}},
            },
        },
    }
}
```

`PositionalArgs` maps the first CLI argument to the `add` field in `MsgAddRequest`, so `add 5` works instead of `add --add 5`.

---

## Step 9 — Wire into app.go

Open `app.go`. You need to make the following additions.

### Imports

```go
counter       "github.com/cosmos/example/x/counter"
counterkeeper "github.com/cosmos/example/x/counter/keeper"
countertypes  "github.com/cosmos/example/x/counter/types"
```

### Module account permissions

```go
maccPerms = map[string][]string{
    // existing entries...
    countertypes.ModuleName: nil,
}
```

### Keeper Field

```go
type ExampleApp struct {
    // existing fields...
    CounterKeeper *counterkeeper.Keeper
}
```

### Store Key

```go
keys := sdk.NewKVStoreKeys(
    // existing keys...
    countertypes.StoreKey,
)
```

### Keeper Instantiation

```go
app.CounterKeeper = counterkeeper.NewKeeper(
    runtime.NewKVStoreService(keys[countertypes.StoreKey]),
    appCodec,
)
```

### Module Manager

```go
app.ModuleManager = module.NewManager(
    // existing modules...
    counter.NewAppModule(appCodec, app.CounterKeeper),
)
```

### BeginBlockers, EndBlockers, and Genesis Order

Add `countertypes.ModuleName` to each of `SetOrderBeginBlockers`, `SetOrderEndBlockers`, `genesisModuleOrder`, and `exportModuleOrder`. Position doesn't matter for this module — place it before `genutiltypes.ModuleName` in each.

---

## Step 10 — Build

```bash
go build ./...
```

Fix any compilation errors before continuing.

---

## Step 11 — Smoke Test

Install the binary and start the chain:

```bash
make install
make start
```

In a second terminal, submit an add transaction:

```bash
exampled tx counter add 5 --from alice --chain-id demo --yes
```

Query the counter:

```bash
exampled query counter count
```

```text
count: "5"
```

Your minimal counter module is working end-to-end.

---

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

---

Next: [Production Counter Walkthrough →](./tutorial-03-counter-walkthrough.md)
