package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/example/x/counter/types"
)

type queryServer struct {
	*Keeper
}

func NewQueryServer(k *Keeper) types.QueryServer {
	return &queryServer{k}
}

func (q queryServer) Count(ctx context.Context, request *types.QueryCountRequest) (*types.QueryCountResponse, error) {
	count, err := q.counter.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return &types.QueryCountResponse{Count: count}, nil
}
