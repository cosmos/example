package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/collections"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if request.GetAdd() >= math.MaxUint64 {
		return nil, ErrNumTooLarge
	}

	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	if params.MaxAddValue > 0 && request.GetAdd() > params.MaxAddValue {
		return nil, ErrExceedsMaxAdd
	}

	// Charge the user if add cost is set
	if !params.AddCost.IsZero() {
		senderAddr, err := sdk.AccAddressFromBech32(request.Sender)
		if err != nil {
			return nil, err
		}
		if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, params.AddCost); err != nil {
			return nil, sdkerrors.Wrap(ErrInsufficientFunds, err.Error())
		}
	}

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

	countMetric.Add(ctx, int64(request.GetAdd()))

	return &types.MsgAddResponse{UpdatedCount: newCount}, nil
}

// UpdateParams updates the module parameters.
func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.authority != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := m.setParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
