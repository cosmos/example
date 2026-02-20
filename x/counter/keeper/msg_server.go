package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/example/x/counter/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper}
}

func (m msgServer) Add(ctx context.Context, request *types.MsgAddRequest) (*types.MsgAddResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	count, err := m.counter.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}

	newCount := count + request.GetAdd()

	err = m.counter.Set(ctx, newCount)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"count_increased",
			sdk.NewAttribute("count", fmt.Sprintf("%v", newCount)),
		),
	)

	return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}
