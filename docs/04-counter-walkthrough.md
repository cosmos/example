# Full Counter Module Walkthrough

If you came here from the module building tutorial, switch back to the `main` branch of the [`cosmos/example` repo](https://github.com/cosmos/example) first:

```bash
git checkout main
```

The minimal counter you built in the previous tutorial captures the core SDK module pattern. The full `x/counter` module example in `main` follows the same pattern and adds several features on top.

This walkthrough is meant to show you exactly what each feature is, what it does, and how you can add a similar feature to any module.

## Minimal vs full counter

The full counter in the `main` branch adds quite a bit of functionality to the minimal tutorial counter.

| Feature | minimal x/counter | full x/counter |
|---|---|---|
| [State](#params-and-authority) | `count` | `count` + `params` |
| [Messages](#params-and-authority) | `Add` | `Add` + `UpdateParams` |
| [Queries](#params-and-authority) | `Count` | `Count` + `Params` |
| [Validation](#expected-keepers-and-fee-collection) | None | `MaxAddValue` limit, overflow check |
| [Fees](#expected-keepers-and-fee-collection) | None | `AddCost` charged via bank module |
| [Authority](#params-and-authority) | None | Governance-gated param updates |
| [Errors](#sentinel-errors) | Generic | Named sentinel errors |
| [Telemetry](#telemetry) | None | OpenTelemetry counter metric |
| [CLI](#autocli) | AutoCLI | AutoCLI + `EnhanceCustomCommand` |
| [Simulation](#simulation) | None | `simsx` weighted operations |
| [Block hooks](#beginblock-and-endblock) | None | `BeginBlock` + `EndBlock` |
| [Unit tests](#unit-tests) | None | Full keeper/msg/query test suite |

The wiring code in [`msg_server.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/msg_server.go), [`query_server.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/query_server.go), [`module.go`](https://github.com/cosmos/example/blob/main/x/counter/module.go), and [`types/`](https://github.com/cosmos/example/tree/main/x/counter/types) is structurally similar between the two. Much of the new keeper logic lives in a single method: `AddCount` in [`keeper.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper.go).

## Params and authority

A [module param](https://docs.cosmos.network/sdk/next/learn/concepts/modules#params) is on-chain configuration that controls how the module behaves without changing the code.

The full counter adds a `Params` type that lets the chain governance configure the module's behavior at runtime. In the full module, params control how large an `Add` can be and how much it costs.

### Where the code lives

- [`proto/example/counter/v1/state.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/state.proto) defines the `Params` type
- [`proto/example/counter/v1/tx.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/tx.proto) adds the `UpdateParams` message
- [`proto/example/counter/v1/query.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/query.proto) adds the `Params` query
- [`x/counter/keeper/keeper.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper.go) stores the params and authority
- [`x/counter/keeper/msg_server.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/msg_server.go) checks the authority on updates
- [`x/counter/keeper/query_server.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/query_server.go) returns the current params

### Try it

You can inspect the current params with:

```bash
exampled query counter params
```

### Add this to your module

To add runtime-configurable params to your own module, make these changes:

1. Define a `Params` type in proto
2. Add a privileged `UpdateParams` message
3. Add a query to read the current params
4. Store the params and authority in your keeper
5. Check the authority in `MsgServer` before writing new params

### state.proto

The relevant addition in [`state.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/state.proto) is:

```proto
message Params {
  uint64 max_add_value = 1;
  repeated cosmos.base.v1beta1.Coin add_cost = 2 [
    (gogoproto.nullable) = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty) = true
  ];
}
```

`MaxAddValue` caps how much a single `Add` call can increment the counter. `AddCost` sets an optional fee charged for each add operation.

### tx.proto - UpdateParams

The relevant addition in [`tx.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/tx.proto) is:

```proto
rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);

message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  Params params = 2 [(gogoproto.nullable) = false];
}

message MsgUpdateParamsResponse {}
```

`UpdateParams` is a privileged message. Only the `authority` address can call it. By default that address is the governance module account, so params can only be changed through a governance proposal.

### query.proto - Params

[`query.proto`](https://github.com/cosmos/example/blob/main/proto/example/counter/v1/query.proto) adds a second query to expose the current params:

```proto
rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
```

### The authority pattern

The keeper stores the authority address and checks it on every `UpdateParams` call:

```go
type Keeper struct {
    // ...
    // authority is the address capable of executing a MsgUpdateParams message.
    // Typically, this should be the x/gov module account.
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

This pattern, storing authority in the keeper and checking it in `MsgServer`, is the standard Cosmos SDK approach to governance-gated configuration.


## Expected keepers and fee collection

This section shows the standard Cosmos SDK pattern for [module-to-module interaction](https://docs.cosmos.network/sdk/next/learn/concepts/modules#inter-module-access). `x/counter` uses an expected keeper to call into the bank module and charge a fee for each add operation.

### Where the code lives

- [`x/counter/types/expected_keepers.go`](https://github.com/cosmos/example/blob/main/x/counter/types/expected_keepers.go) defines the narrow bank keeper interface
- [`x/counter/keeper/keeper.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper.go) stores the bank keeper dependency and charges the fee in `AddCount`
- [`app.go`](https://github.com/cosmos/example/blob/main/app.go) passes `app.BankKeeper` into `counterkeeper.NewKeeper`
- [`app.go`](https://github.com/cosmos/example/blob/main/app.go) adds a module account entry so the counter module can receive fees

### app.go changes

This feature requires two `app.go` changes:

- add `countertypes.ModuleName: nil` to `maccPerms`
- pass `app.BankKeeper` into `counterkeeper.NewKeeper(...)`

In [`app.go`](https://github.com/cosmos/example/blob/main/app.go), those changes look like this:

```go
maccPerms = map[string][]string{
    // ...
    countertypes.ModuleName: nil,
}
```

```go
app.CounterKeeper = counterkeeper.NewKeeper(
    runtime.NewKVStoreService(keys[countertypes.StoreKey]),
    appCodec,
    app.BankKeeper,
)
```

### Try it

Submit an add transaction and the configured `AddCost` fee will be charged from the sender:

```bash
exampled tx counter add 5 --from alice --chain-id demo --yes
```

### Add this to your module

To add fee collection through the bank module, make these changes:

1. Define a narrow bank keeper interface in `types/expected_keepers.go`
2. Add a `bankKeeper` field to your keeper
3. Charge the fee inside your keeper business logic
4. Add a module account entry in `maccPerms`
5. Pass `app.BankKeeper` into your keeper constructor in `app.go`

### expected_keepers.go

Rather than importing the bank module directly, the counter module defines the minimal interface it needs:

```go
// x/counter/types/expected_keepers.go
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
    bankKeeper types.BankKeeper
    authority  string
}
```

### Fee charging in AddCount

```go
func (k *Keeper) AddCount(ctx context.Context, sender string, amount uint64) (uint64, error) {
    if amount >= math.MaxUint64 {
        return 0, ErrNumTooLarge
    }

    params, err := k.GetParams(ctx)
    if err != nil {
        return 0, err
    }

    if params.MaxAddValue > 0 && amount > params.MaxAddValue {
        return 0, ErrExceedsMaxAdd
    }

    if !params.AddCost.IsZero() {
        senderAddr, err := sdk.AccAddressFromBech32(sender)
        if err != nil {
            return 0, err
        }
        if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, params.AddCost); err != nil {
            return 0, sdkerrors.Wrap(ErrInsufficientFunds, err.Error())
        }
    }

    count, err := k.GetCount(ctx)
    if err != nil {
        return 0, err
    }

    newCount := count + amount
    if err := k.counter.Set(ctx, newCount); err != nil {
        return 0, err
    }

    sdkCtx := sdk.UnwrapSDKContext(ctx)
    sdkCtx.EventManager().EmitEvent(
        sdk.NewEvent(
            "count_increased",
            sdk.NewAttribute("count", fmt.Sprintf("%v", newCount)),
        ),
    )

    countMetric.Add(ctx, int64(amount))

    return newCount, nil
}
```

All the business logic, validation, fee charging, state mutation, events, and telemetry, lives in `AddCount`. The `MsgServer` stays thin:

```go
func (m msgServer) Add(ctx context.Context, req *types.MsgAddRequest) (*types.MsgAddResponse, error) {
    newCount, err := m.AddCount(ctx, req.GetSender(), req.GetAdd())
    if err != nil {
        return nil, err
    }
    return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}
```

Because `AddCount` is a named keeper method, it can also be called from `BeginBlock`, governance hooks, or other modules, not just from the `MsgServer`.

### Module accounts

A module account is an on-chain account owned by a module instead of a user. Modules use module accounts to hold funds, receive fees, or get special permissions like minting or burning.

Because `x/counter` receives fees from users, it needs a module account entry in [`app.go`](https://github.com/cosmos/example/blob/main/app.go):

```go
maccPerms = map[string][]string{
    // ...
    countertypes.ModuleName: nil,
}
```

This lives in the `maccPerms` map in [`app.go`](https://github.com/cosmos/example/blob/main/app.go). Here, `nil` means the module account can receive funds but does not get extra permissions like minting or burning.

## Sentinel errors

Rather than returning generic errors, `x/counter` defines named sentinel errors with registered codes. That makes failures easier to understand and easier for clients to match on programmatically.

### Where the code lives

- [`x/counter/keeper/errors.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/errors.go) defines the registered module errors
- [`x/counter/keeper/keeper.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper.go) returns those errors from business logic checks

```go
// keeper/errors.go
var (
    ErrNumTooLarge       = errors.Register("counter", 0, "requested integer to add is too large")
    ErrExceedsMaxAdd     = errors.Register("counter", 1, "add value exceeds max allowed")
    ErrInsufficientFunds = errors.Register("counter", 2, "insufficient funds to pay add cost")
)
```

Registered errors produce structured error responses on-chain that clients can match against by code, not just by string. Each error code must be unique within the module and greater than zero (code `1` is reserved for internal SDK errors). To check whether an error is of a specific sentinel type, use `errors.Is(err, ErrInsufficientFunds)` — this works correctly even when the error has been wrapped with additional context via `errorsmod.Wrap` or `errorsmod.Wrapf`.

All validation — both stateless field checks and stateful business logic checks — should live in the `msgServer` method or the keeper function it calls. The older `ValidateBasic` method on message types is deprecated: prefer performing all validation inside the message server. If your message type does implement `ValidateBasic`, the SDK still calls it for backward compatibility, but new modules should not rely on it.

## Telemetry

[Telemetry](https://docs.cosmos.network/sdk/next/guides/testing/telemetry) records how often the counter is updated so you can observe module activity in an OpenTelemetry-compatible system.

### Where the code lives

- [`x/counter/keeper/telemetry.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/telemetry.go) defines the meter and counter metric
- [`x/counter/keeper/keeper.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper.go) records the metric from `AddCount`

```go
// x/counter/keeper/telemetry.go
var (
    meter = otel.Meter("github.com/cosmos/example/x/counter")

    countMetric metric.Int64Counter
)

func init() {
    var err error
    countMetric, err = meter.Int64Counter("count")
    if err != nil {
        panic(err)
    }
}
```

`countMetric.Add(ctx, int64(amount))` in `AddCount` increments an OpenTelemetry counter every time the module state is updated. This makes module activity visible in any OTel-compatible observability system.

## AutoCLI

[AutoCLI](https://docs.cosmos.network/sdk/next/guides/tooling/autocli) exposes the module's queries and transactions as CLI commands. The full module example keeps the same basic AutoCLI setup as the minimal module and adds the recommended setting for custom command integration.

### Where the code lives

- [`x/counter/autocli.go`](https://github.com/cosmos/example/blob/main/x/counter/autocli.go) defines the generated query and tx commands

### Try it

These commands come from the AutoCLI configuration. `count` and `add` are customized explicitly in `autocli.go`, and `params` is still available from the generated query service.

```bash
exampled query counter count
exampled query counter params
exampled tx counter add 5 --from alice --chain-id demo --yes
```

Both modules use AutoCLI. The only difference is that `x/counter` sets `EnhanceCustomCommand: true`, which merges any hand-written CLI commands with the auto-generated ones. Since neither module has hand-written commands, it is a no-op here, but it is a good default for fuller modules.

The [`autocli.go`](https://github.com/cosmos/example/blob/main/x/counter/autocli.go) file in `x/counter`:

```go
// autocli.go
func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
    return &autocliv1.ModuleOptions{
        Query: &autocliv1.ServiceCommandDescriptor{
            Service:              "example.counter.Query",
            EnhanceCustomCommand: true,
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                {RpcMethod: "Count", Use: "count", Short: "Query the current counter value"},
            },
        },
        Tx: &autocliv1.ServiceCommandDescriptor{
            Service:              "example.counter.Msg",
            EnhanceCustomCommand: true,
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                {RpcMethod: "Add", Use: "add [amount]", Short: "Add to the counter",
                    PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "add"}}},
            },
        },
    }
}
```


## Simulation

[Simulation](https://docs.cosmos.network/sdk/next/guides/testing/simulator) lets the SDK generate randomized transactions against the module during fuzz-style testing.

### Where the code lives

- [`x/counter/simulation/msg_factory.go`](https://github.com/cosmos/example/blob/main/x/counter/simulation/msg_factory.go) defines how to generate random `Add` messages
- [`x/counter/module.go`](https://github.com/cosmos/example/blob/main/x/counter/module.go) registers those weighted operations

### Test it

You can exercise simulation through the repo's simulation test targets described in the running and testing tutorial.

`x/counter` implements `simsx`-based simulation, which lets the SDK's simulation framework generate random `Add` transactions during fuzz testing:

```go
// x/counter/simulation/msg_factory.go
func MsgAddFactory() simsx.SimMsgFactoryFn[*types.MsgAddRequest] {
    return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *types.MsgAddRequest) {
        sender := testData.AnyAccount(reporter)
        if reporter.IsSkipped() {
            return nil, nil
        }

        r := testData.Rand()
        addAmount := uint64(r.Intn(100) + 1)

        msg := &types.MsgAddRequest{
            Sender: sender.AddressBech32,
            Add:    addAmount,
        }

        return []simsx.SimAccount{sender}, msg
    }
}
```

`module.go` registers this factory:

```go
func (a AppModule) WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry) {
    reg.Add(weights.Get("msg_add", 100), simulation.MsgAddFactory())
}
```


## BeginBlock and EndBlock

These [hooks](https://docs.cosmos.network/sdk/next/learn/concepts/modules#block-hooks) let a module run code automatically at the start or end of every block. In `x/counter`, they are purposefully empty to demonstrate where and how these features can be added.

### Where the code lives

- [`x/counter/module.go`](https://github.com/cosmos/example/blob/main/x/counter/module.go) implements `BeginBlock` and `EndBlock`
- [`app.go`](https://github.com/cosmos/example/blob/main/app.go) adds the module to `SetOrderBeginBlockers` and `SetOrderEndBlockers`

### app.go changes

Because the module advertises block hooks, [`app.go`](https://github.com/cosmos/example/blob/main/app.go) must include `countertypes.ModuleName` in both blocker order lists.

### Add this to your module

To add begin and end blockers to your own module, make two changes:

1. Implement the hooks in `x/<module>/module.go`
2. Add your module name to `SetOrderBeginBlockers` and `SetOrderEndBlockers` in `app.go`

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

In [`app.go`](https://github.com/cosmos/example/blob/main/app.go), the module is added to the blocker order lists like this:

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

`x/counter` has no per-block logic, so both methods return nil. They exist to demonstrate the pattern: modules that need per-block execution (staking, distribution) implement real logic here. For example, a counter that auto-increments every block would call `k.AddCount(ctx, 1)` from `BeginBlock` instead of exposing a message type.

## Unit tests

The full module example includes a real [test suite](https://docs.cosmos.network/sdk/next/learn/concepts/testing) for keeper logic, query behavior, message handling, and bank keeper interactions.

### Where the code lives

- [`x/counter/keeper/keeper_test.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper_test.go)
- [`x/counter/keeper/msg_server_test.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/msg_server_test.go)
- [`x/counter/keeper/query_server_test.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/query_server_test.go)

### Run them

You can run the counter module tests directly with:

```bash
go test ./x/counter/...
```

### Add this to your module

Start with keeper, message server, and query server tests. If your module depends on another keeper, use a small mock interface like `MockBankKeeper` so you can control success and failure cases in isolation.

`x/counter` ships a full test suite in [`x/counter/keeper/`](https://github.com/cosmos/example/tree/main/x/counter/keeper):

| File | What it tests |
|---|---|
| `keeper_test.go` | `KeeperTestSuite` setup, `InitGenesis`, `ExportGenesis`, `GetCount`, `AddCount`, `SetParams` |
| `msg_server_test.go` | `MsgAdd`, event emission, `MsgUpdateParams` |
| `query_server_test.go` | `QueryCount`, `QueryParams` |

All three files share the `KeeperTestSuite` struct defined in [`keeper_test.go`](https://github.com/cosmos/example/blob/main/x/counter/keeper/keeper_test.go), which sets up an isolated in-memory store, a mock bank keeper, and a real keeper instance:

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

## Gas

`minimum-gas-prices` in `app.toml` sets the minimum fee a node requires before it will accept and relay a transaction. The local dev chain started by `make start` leaves this empty, so transactions are accepted with no fee beyond the `AddCost` module parameter.

To require a minimum network fee, set it in `app.toml`:

```toml
minimum-gas-prices = "0.025stake"
```

Transactions that don't meet the minimum will be rejected by the node before they reach your module. This is a per-node setting, not a chain-wide consensus rule, so validators on a live network each configure their own threshold.

Next: [Running and Testing →](./05-run-and-test.md)
