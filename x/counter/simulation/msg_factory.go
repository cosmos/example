package simulation

import (
	"context"

	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	"github.com/cosmos/example/x/counter/types"
)

// MsgAddFactory creates a simulation message factory for MsgAddRequest.
func MsgAddFactory() simsx.SimMsgFactoryFn[*types.MsgAddRequest] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *types.MsgAddRequest) {
		sender := testData.AnyAccount(reporter)
		if reporter.IsSkipped() {
			return nil, nil
		}

		r := testData.Rand()
		addAmount := uint64(r.Intn(100) + 1)

		msg := &types.MsgAddRequest{
			Sender: sender.AddressBech32,
			Add:    addAmount,
		}

		return []simsx.SimAccount{sender}, msg
	}
}
