package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/example/x/counter/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper}
}

func (m msgServer) Add(ctx context.Context, request *types.MsgAddRequest) (*types.MsgAddResponse, error) {
	newCount, err := m.AddCount(ctx, request.GetSender(), request.GetAdd())
	if err != nil {
		return nil, err
	}

	return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}

// UpdateParams updates the module parameters.
func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.authority != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
