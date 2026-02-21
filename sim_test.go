//go:build sims

package example

import (
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simsx "github.com/cosmos/cosmos-sdk/testutil/simsx"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

func init() {
	simcli.GetSimulatorFlags()
}

func TestFullAppSimulation(t *testing.T) {
	simsx.Run(t, NewExampleApp, setupStateFactory)
}

func setupStateFactory(app *ExampleApp) simsx.SimStateFactory {
	return simsx.SimStateFactory{
		Codec:         app.AppCodec(),
		AppStateFn:    simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		BlockedAddr:   BlockedAddresses(),
		AccountSource: app.AccountKeeper,
		BalanceSource: app.BankKeeper,
	}
}

func TestAppStateDeterminism(t *testing.T) {
	simsx.Run(t, NewExampleApp, setupStateFactory, func(tb testing.TB, ti simsx.TestInstance[*ExampleApp], accs []simtypes.Account) {
		tb.Log("running determinism test...")
	})
}
