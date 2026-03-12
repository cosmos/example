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

func (q queryServer) Count(ctx context.Context, _ *types.QueryCountRequest) (*types.QueryCountResponse, error) {
	count, err := q.GetCount(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCountResponse{Count: count}, nil
}

func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.params.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}
