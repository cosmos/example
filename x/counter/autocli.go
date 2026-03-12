package counter

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              "example.counter.Query",
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Count",
					Use:       "count",
					Short:     "Query the current counter value",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              "example.counter.Msg",
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Add",
					Use:            "add [amount]",
					Short:          "Add to the counter",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "add"}},
				},
			},
		},
	}
}
