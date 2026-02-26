package tests

import (
	"fmt"
	"os"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	pruningtypes "cosmossdk.io/store/pruning/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	example "github.com/cosmos/example"
)

// NewTestNetworkFixture returns a new ExampleApp network fixture for e2e tests.
func NewTestNetworkFixture() network.TestFixture {
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)

	app := example.NewExampleApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(dir))

	appCtr := func(val network.ValidatorI) servertypes.Application {
		return example.NewExampleApp(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			bam.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
		)
	}

	return network.TestFixture{
		AppConstructor: appCtr,
		GenesisState:   app.DefaultGenesis(),
		EncodingConfig: testutil.TestEncodingConfig{
			InterfaceRegistry: app.InterfaceRegistry(),
			Codec:             app.AppCodec(),
			TxConfig:          app.TxConfig(),
			Amino:             app.LegacyAmino(),
		},
	}
}
