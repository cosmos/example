package dex

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              "example.dex.Query",
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Pool",
					Use:       "pool [pool-id]",
					Short:     "Get AMM pool by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              "example.dex.Msg",
			EnhanceCustomCommand: false,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Swap",
					Use:       "swap [pool-id] [amount-in] [min-amount-out] [denom-in]",
					Short:     "Swap tokens in a pool",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "amount_in"},
						{ProtoField: "min_amount_out"},
						{ProtoField: "denom_in"},
					},
				},
			},
		},
	}
}
