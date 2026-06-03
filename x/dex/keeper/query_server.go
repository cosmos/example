package keeper

import (
	"context"

	"github.com/cosmos/example/x/dex/types"
)

type queryServer struct {
	*Keeper
}

func NewQueryServer(k *Keeper) types.QueryServer {
	return &queryServer{k}
}

func (q queryServer) Pool(ctx context.Context, request *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	pool, err := q.GetPool(ctx, request.PoolId)
	if err != nil {
		return nil, err
	}
	return &types.QueryPoolResponse{Pool: &pool}, nil
}
