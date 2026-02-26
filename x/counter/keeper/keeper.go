package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
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

// setParams sets the module parameters.
func (k *Keeper) setParams(ctx context.Context, params types.Params) error {
	return k.params.Set(ctx, params)
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

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}
