package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/example/x/dex/types"
)

type Keeper struct {
	Schema     collections.Schema
	pools      collections.Map[string, types.Pool]
	bankKeeper types.BankKeeper
}

func NewKeeper(storeService store.KVStoreService, cdc codec.Codec, bankKeeper types.BankKeeper) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		pools: collections.NewMap(
			sb,
			collections.NewPrefix(0),
			"pools",
			collections.StringKey,
			codec.CollValue[types.Pool](cdc),
		),
		bankKeeper: bankKeeper,
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return &k
}

func (k *Keeper) GetPool(ctx context.Context, poolID string) (types.Pool, error) {
	return k.pools.Get(ctx, poolID)
}

func (k *Keeper) SetPool(ctx context.Context, pool types.Pool) error {
	return k.pools.Set(ctx, pool.Id, pool)
}

func (k *Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	for _, pool := range gs.Pools {
		if err := k.SetPool(ctx, pool); err != nil {
			return err
		}
		reserveA, err := sdkmath.ParseUint(pool.ReserveA)
		if err != nil {
			return err
		}
		reserveB, err := sdkmath.ParseUint(pool.ReserveB)
		if err != nil {
			return err
		}
		coins := sdk.NewCoins(
			sdk.NewCoin(pool.DenomA, sdkmath.Int(reserveA)),
			sdk.NewCoin(pool.DenomB, sdkmath.Int(reserveB)),
		)
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var pools []types.Pool
	err := k.pools.Walk(ctx, nil, func(_ string, pool types.Pool) (bool, error) {
		pools = append(pools, pool)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Pools: pools}, nil
}

func (k *Keeper) Swap(ctx context.Context, msg *types.MsgSwap) (string, error) {
	pool, err := k.GetPool(ctx, msg.PoolId)
	if err != nil {
		return "", err
	}
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return "", err
	}
	amountInInt, err := sdkmath.ParseUint(msg.AmountIn)
	if err != nil {
		return "", err
	}
	denomIn := msg.DenomIn
	if denomIn == "" {
		denomIn = pool.DenomA
	}
	var amountOut string
	switch denomIn {
	case pool.DenomA:
		amountOut, err = calcAmountOut(ctx, pool.ReserveA, pool.ReserveB, msg.AmountIn, msg.MinAmountOut)
		if err != nil {
			return "", err
		}

		amountOutInt, err := sdkmath.ParseUint(amountOut)
		if err != nil {
			return "", err
		}
		reserveBInt, err := sdkmath.ParseUint(pool.ReserveB)
		if err != nil {
			return "", err
		}
		if amountOutInt.GT(reserveBInt) {
			return "", errors.New("insufficient pool liquidity")
		}

		coinsIn := sdk.NewCoins(sdk.NewCoin(pool.DenomA, sdkmath.Int(amountInInt)))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coinsIn); err != nil {
			return "", err
		}

		coinsOut := sdk.NewCoins(sdk.NewCoin(pool.DenomB, sdkmath.Int(amountOutInt)))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, coinsOut); err != nil {
			return "", err
		}
		reserveAInt, err := sdkmath.ParseUint(pool.ReserveA)
		if err != nil {
			return "", err
		}
		reserveAInt = reserveAInt.Add(amountInInt)
		pool.ReserveA = reserveAInt.String()

		pool.ReserveB = reserveBInt.Sub(amountOutInt).String()
		if err := k.SetPool(ctx, pool); err != nil {
			return "", err
		}
	case pool.DenomB:
		amountOut, err = calcAmountOut(ctx, pool.ReserveB, pool.ReserveA, msg.AmountIn, msg.MinAmountOut)
		if err != nil {
			return "", err
		}

		reserveAInt, err := sdkmath.ParseUint(pool.ReserveA)
		if err != nil {
			return "", err
		}
		amountOutInt, err := sdkmath.ParseUint(amountOut)
		if err != nil {
			return "", err
		}
		if amountOutInt.GT(reserveAInt) {
			return "", errors.New("insufficient pool liquidity")
		}
		coinsIn := sdk.NewCoins(sdk.NewCoin(pool.DenomB, sdkmath.Int(amountInInt)))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coinsIn); err != nil {
			return "", err
		}

		coinsOut := sdk.NewCoins(sdk.NewCoin(pool.DenomA, sdkmath.Int(amountOutInt)))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, coinsOut); err != nil {
			return "", err
		}
		reserveBInt, err := sdkmath.ParseUint(pool.ReserveB)
		if err != nil {
			return "", err
		}
		reserveBInt = reserveBInt.Add(amountInInt)
		pool.ReserveB = reserveBInt.String()

		pool.ReserveA = reserveAInt.Sub(amountOutInt).String()
		if err := k.SetPool(ctx, pool); err != nil {
			return "", err
		}
	default:
		return "", errors.New("invalid denom_in")
	}

	return amountOut, nil
}

func calcAmountOut(_ context.Context, reserveIn, reserveOut, amountIn, minOut string) (string, error) {
	reserveInInt, err := sdkmath.ParseUint(reserveIn)
	if err != nil {
		return "", err
	}

	if reserveInInt.IsZero() {
		return "", errors.New("reserveIn is zero")
	}

	reserveOutInt, err := sdkmath.ParseUint(reserveOut)
	if err != nil {
		return "", err
	}
	amountInInt, err := sdkmath.ParseUint(amountIn)
	if err != nil {
		return "", err
	}

	a := reserveOutInt.Mul(amountInInt)
	b := reserveInInt.Add(amountInInt)
	result := a.Quo(b)

	if result.IsZero() {
		return "", errors.New("result is zero")
	}

	minOutInt, err := sdkmath.ParseUint(minOut)
	if err != nil {
		return "", err
	}
	if result.LT(minOutInt) {
		return "", errors.New("result is less than minOut")
	}

	return result.String(), nil
}
