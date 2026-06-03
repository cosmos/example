package keeper

import (
	"context"

	"github.com/cosmos/example/x/dex/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{k}
}

func (m msgServer) Swap(ctx context.Context, msg *types.MsgSwap) (*types.MsgSwapResponse, error) {
	amountOut, err := m.Keeper.Swap(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSwapResponse{AmountOut: amountOut}, nil
}
