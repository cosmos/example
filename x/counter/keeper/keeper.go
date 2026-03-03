package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/example/x/counter/types"
)

type Keeper struct {
	Schema     collections.Schema
	counter    collections.Item[uint64]
	params     collections.Item[types.Params]
	bankKeeper types.BankKeeper

	// authority is the address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority string
}

type Options func(k *Keeper)

// WithAuthority sets a custom authority on the module. This allows developers to set accounts other than the
// governance module to control this module's params.
func WithAuthority(authority string) Options {
	return func(k *Keeper) {
		k.authority = authority
	}
}

func NewKeeper(storeService store.KVStoreService, cdc codec.Codec, bankKeeper types.BankKeeper, opts ...Options) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		counter:    collections.NewItem(sb, collections.NewPrefix(0), "counter", collections.Uint64Value),
		params:     collections.NewItem(sb, collections.NewPrefix(1), "params", codec.CollValue[types.Params](cdc)),
		bankKeeper: bankKeeper,
		authority:  authtypes.NewModuleAddress(govtypes.ModuleName).String(), // default authority is governance module.
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	for _, opt := range opts {
		opt(&k)
	}

	return &k
}

// GetParams returns the module parameters.
func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

// SetParams sets the module parameters.
func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

// GetCount returns the current counter value.
func (k *Keeper) GetCount(ctx context.Context) (uint64, error) {
	count, err := k.counter.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return 0, err
	}
	return count, nil
}

// AddCount adds the specified amount to the counter after validating against params
// and charging any applicable fees. Returns the new count.
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

	// Charge the user if add cost is set
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

func (k *Keeper) InitGenesis(ctx context.Context, genesis *types.GenesisState) error {
	if err := k.counter.Set(ctx, genesis.Count); err != nil {
		return err
	}
	return k.params.Set(ctx, genesis.Params)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	count, err := k.counter.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	params, err := k.params.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return &types.GenesisState{Count: count, Params: params}, nil
}
