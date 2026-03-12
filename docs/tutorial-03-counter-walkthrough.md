# Production Counter Walkthrough

The minimal counter you built in the previous tutorial captures the core SDK module pattern. The `x/counter` module in this repository follows the same pattern — it just adds several production features on top.

This walkthrough explains what was added, where it lives, and why.

The wiring code — `msg_server.go`, `query_server.go`, `module.go`, `types/` — is structurally identical between the two. Almost all of the new logic lives in a single method: `keeper.AddCount`.

---

## Minimal vs production

| Feature | Minimal | x/counter |
|---|---|---|
| State | `count` | `count` + `params` |
| Messages | `Add` | `Add` + `UpdateParams` |
| Queries | `Count` | `Count` + `Params` |
| Validation | None | `MaxAddValue` limit, overflow check |
| Fees | None | `AddCost` charged via bank module |
| Authority | None | Governance-gated param updates |
| Errors | Generic | Named sentinel errors |
| Telemetry | None | OpenTelemetry counter metric |
| CLI | AutoCLI | AutoCLI + `EnhanceCustomCommand` |
| Simulation | None | `simsx` weighted operations |
| Block hooks | None | `BeginBlock` + `EndBlock` stubs |
| Unit tests | None | Full keeper/msg/query test suite |

---

## Params and authority

The most significant addition is a `Params` type that lets the chain governance configure the module's behavior at runtime.

### state.proto

A new proto file defines the `Params` type:

```proto
// proto/example/counter/v1/state.proto
message Params {
  uint64 max_add_value = 1;
  repeated cosmos.base.v1beta1.Coin add_cost = 2;
}
```

`MaxAddValue` caps how much a single `Add` call can increment the counter. `AddCost` sets an optional fee charged for each add operation.

### tx.proto — UpdateParams

A second message is added to `tx.proto`:

```proto
rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);

message MsgUpdateParams {
  string authority = 1;
  Params params    = 2;
}
```

`UpdateParams` is a privileged message — only the `authority` address can call it. By default that address is the governance module account, so params can only be changed through a governance proposal.

### query.proto — Params

A second query is added to expose the current params:

```proto
rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
```

### The authority pattern

The keeper stores the authority address and checks it on every `UpdateParams` call:

```go
// keeper.go
type Keeper struct {
    // ...
    authority string
}
```

```go
// msg_server.go
func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
    if m.authority != msg.Authority {
        return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner,
            "invalid authority; expected %s, got %s", m.authority, msg.Authority)
    }
    return &types.MsgUpdateParamsResponse{}, m.SetParams(ctx, msg.Params)
}
```

The authority defaults to the governance module account at keeper construction:

```go
authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
```

This pattern — storing authority in the keeper, checking it in `MsgServer` — is the standard Cosmos SDK approach to governance-gated configuration.

---

## Fee collection — BankKeeper

`x/counter` charges a fee for each add operation when `AddCost` is set. This requires calling into the bank module.

### expected_keepers.go

Rather than importing the bank module directly, the counter module defines the minimal interface it needs:

```go
// types/expected_keepers.go
type BankKeeper interface {
    SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
```

This keeps the dependency explicit and narrow. The counter module cannot accidentally call any other bank method.

### Keeper struct

```go
type Keeper struct {
    Schema     collections.Schema
    counter    collections.Item[uint64]
    params     collections.Item[types.Params]
    bankKeeper types.BankKeeper  // injected at construction
    authority  string
}
```

### Fee charging in AddCount

```go
func (k *Keeper) AddCount(ctx context.Context, sender string, amount uint64) (uint64, error) {
    // overflow check
    if amount >= math.MaxUint64 {
        return 0, ErrNumTooLarge
    }

    params, err := k.GetParams(ctx)

    // enforce MaxAddValue
    if params.MaxAddValue > 0 && amount > params.MaxAddValue {
        return 0, ErrExceedsMaxAdd
    }

    // charge fee if configured
    if !params.AddCost.IsZero() {
        senderAddr, _ := sdk.AccAddressFromBech32(sender)
        if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, params.AddCost); err != nil {
            return 0, sdkerrors.Wrap(ErrInsufficientFunds, err.Error())
        }
    }

    // update state
    count, _ := k.GetCount(ctx)
    newCount := count + amount
    k.counter.Set(ctx, newCount)

    // emit event
    sdkCtx.EventManager().EmitEvent(sdk.NewEvent("count_increased",
        sdk.NewAttribute("count", fmt.Sprintf("%v", newCount)),
    ))

    countMetric.Add(ctx, int64(amount))
    return newCount, nil
}
```

All the business logic — validation, fee charging, state mutation, events, and telemetry — lives in `AddCount`. The `MsgServer` stays thin:

```go
func (m msgServer) Add(ctx context.Context, req *types.MsgAddRequest) (*types.MsgAddResponse, error) {
    newCount, err := m.AddCount(ctx, req.GetSender(), req.GetAdd())
    if err != nil {
        return nil, err
    }
    return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}
```

Because `AddCount` is a named keeper method, it can also be called from `BeginBlock`, governance hooks, or other modules — not just from the `MsgServer`.

### Module account

Because the counter module receives fees from users, it needs a module account in the bank module's permission map. This is registered in `app.go`:

```go
maccPerms = map[string][]string{
    // ...
    countertypes.ModuleName: nil,
}
```

`nil` means the module account can receive funds but not mint or burn them.

---

## Sentinel errors

Rather than returning generic errors, `x/counter` defines named errors with registered codes:

```go
// keeper/errors.go
var (
    ErrNumTooLarge       = errors.Register("counter", 0, "requested integer to add is too large")
    ErrExceedsMaxAdd     = errors.Register("counter", 1, "add value exceeds max allowed")
    ErrInsufficientFunds = errors.Register("counter", 2, "insufficient funds to pay add cost")
)
```

Registered errors produce structured error responses on-chain that clients can match against by code, not just by string.

---

## Telemetry

```go
// keeper/telemetry.go
var (
    meter       = otel.Meter("github.com/cosmos/example/x/counter")
    countMetric metric.Int64Counter
)

func init() {
    countMetric, err = meter.Int64Counter("count")
}
```

`countMetric.Add(ctx, int64(amount))` in `AddCount` increments an OpenTelemetry counter every time the module state is updated. This makes module activity visible in any OTel-compatible observability system.

---

## AutoCLI

Both modules use AutoCLI. The only difference is that `x/counter` sets `EnhanceCustomCommand: true`, which merges any hand-written CLI commands with the auto-generated ones. Since neither module has hand-written commands, it is a no-op here — but it is the recommended default for production modules.

The `autocli.go` in `x/counter`:

```go
// autocli.go
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

---

## Simulation

`x/counter` implements `simsx`-based simulation, which lets the SDK's simulation framework generate random `Add` transactions during fuzz testing:

```go
// simulation/msg_factory.go
func MsgAddFactory() simsx.SimMsgFactoryFn[*types.MsgAddRequest] {
    return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *types.MsgAddRequest) {
        sender := testData.AnyAccount(reporter)
        addAmount := uint64(testData.Rand().Intn(100) + 1)
        return []simsx.SimAccount{sender}, &types.MsgAddRequest{
            Sender: sender.AddressBech32,
            Add:    addAmount,
        }
    }
}
```

`module.go` registers this factory:

```go
func (a AppModule) WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry) {
    reg.Add(weights.Get("msg_add", 100), simulation.MsgAddFactory())
}
```

---

## BeginBlock and EndBlock

`module.go` implements `HasBeginBlocker` and `HasEndBlocker`:

```go
func (a AppModule) BeginBlock(ctx context.Context) error {
    // optional: logic to execute at the start of every block
    return nil
}

func (a AppModule) EndBlock(ctx context.Context) error {
    // optional: logic to execute at the end of every block
    return nil
}
```

`x/counter` has no per-block logic, so both methods return nil. They exist to demonstrate the pattern: modules that need per-block execution (staking, distribution) implement real logic here. For example, a counter that auto-increments every block would call `k.AddCount(ctx, 1)` from `BeginBlock` instead of exposing a message type.

---

## Unit tests

`x/counter` ships a full test suite in `x/counter/keeper/`:

| File | What it tests |
|---|---|
| `keeper_test.go` | `KeeperTestSuite` setup, `InitGenesis`, `ExportGenesis`, `GetCount`, `AddCount`, `SetParams` |
| `msg_server_test.go` | `MsgAdd`, event emission, `MsgUpdateParams` |
| `query_server_test.go` | `QueryCount`, `QueryParams` |

All three files share the `KeeperTestSuite` struct defined in `keeper_test.go`, which sets up an isolated in-memory store, a mock bank keeper, and a real keeper instance:

```go
type KeeperTestSuite struct {
    suite.Suite
    ctx         sdk.Context
    keeper      *keeper.Keeper
    queryClient types.QueryClient
    msgServer   types.MsgServer
    bankKeeper  *MockBankKeeper
    authority   string
}
```

`MockBankKeeper` lets tests control exactly what the bank keeper returns without needing a real bank module:

```go
type MockBankKeeper struct {
    SendCoinsFromAccountToModuleFn func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
```

Tests set `SendCoinsFromAccountToModuleFn` to simulate success or failure:

```go
s.bankKeeper.SendCoinsFromAccountToModuleFn = func(...) error {
    return errors.New("insufficient funds")
}
```

---

## Additional app.go Wiring

Tutorial 02 already had you add the basic app wiring for the minimal module: imports, the keeper field, the store key, keeper construction, module registration, and genesis/export ordering. The production module adds a few more `app.go` changes on top of that.

### Module account permissions

Because the production module can collect `AddCost` fees, it needs a module account entry in `maccPerms`:

```go
maccPerms = map[string][]string{
    // ...
    countertypes.ModuleName: nil,
}
```

This lives in the top-level `maccPerms` map in `app.go`. `nil` means the account can receive funds but does not get mint or burn permissions.

### Keeper construction with BankKeeper

The production keeper takes one extra dependency: `app.BankKeeper`. This is added where `app.CounterKeeper` is constructed:

```go
app.CounterKeeper = counterkeeper.NewKeeper(
    runtime.NewKVStoreService(keys[countertypes.StoreKey]),
    appCodec,
    app.BankKeeper,
)
```

`app.BankKeeper` satisfies the `types.BankKeeper` interface defined in `expected_keepers.go`, so the counter module can charge fees without importing the full bank keeper type directly.

### Begin and end block ordering

The production `AppModule` implements `BeginBlock` and `EndBlock` in `x/counter/module.go`, even though both methods currently return `nil`. Because the module advertises those hooks, `app.go` also adds `countertypes.ModuleName` to `SetOrderBeginBlockers` and `SetOrderEndBlockers`:

```go
app.ModuleManager.SetOrderBeginBlockers(
    // ...
    countertypes.ModuleName,
)

app.ModuleManager.SetOrderEndBlockers(
    // ...
    countertypes.ModuleName,
)
```

That tells the module manager where the counter module belongs in the per-block execution order. The production branch keeps the same `genesisModuleOrder` and `exportModuleOrder` wiring from Tutorial 02 as well.

---

Next: [Running and Testing →](./tutorial-04-run-and-test.md)


<!-- todo: We need to talk about configuring gas and other configurations. We also need to break up these parts into what the code is and how you wire it into AppDecode to enable it in your chain.  -->