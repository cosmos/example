package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/example/x/counter/types"
)

type State struct {
	foo int
}
type Keeper struct {
	counter collections.Item[uint64]
}

func NewKeeper(storeService store.KVStoreService, cdc codec.Codec) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		counter: collections.NewItem(sb, collections.NewPrefix(0), "counter", collections.Uint64Value),
	}
	_, err := sb.Build()
	if err != nil {
		panic(err)
	}
	return &k
}

func (k *Keeper) InitGenesis(ctx context.Context, genesis *types.GenesisState) error {
	return k.counter.Set(ctx, genesis.Count)
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	count, err := k.counter.Get(ctx)
	if !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return &types.GenesisState{Count: count}, nil
}
