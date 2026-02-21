package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/example/x/counter/types"
)

// RandomizedGenState generates a random GenesisState for the counter module.
func RandomizedGenState(simState *module.SimulationState) {
	var count uint64
	simState.AppParams.GetOrGenerate("counter_count", &count, simState.Rand,
		func(r *rand.Rand) {
			count = uint64(r.Intn(1000))
		},
	)

	counterGenesis := types.GenesisState{
		Count: count,
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&counterGenesis)
}
