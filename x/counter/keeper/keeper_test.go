package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/cosmos/example/x/counter/keeper"
	"github.com/cosmos/example/x/counter/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	keeper      *keeper.Keeper
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey("counter")
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	k := keeper.NewKeeper(storeService, encCfg.Codec)

	s.ctx = ctx
	s.keeper = k

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(k))
	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(k)
}

func (s *KeeperTestSuite) TestInitGenesis() {
	testCases := []struct {
		name    string
		genesis *types.GenesisState
		expErr  bool
	}{
		{
			name:    "default genesis",
			genesis: &types.GenesisState{},
			expErr:  false,
		},
		{
			name:    "custom count",
			genesis: &types.GenesisState{Count: 100},
			expErr:  false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			err := s.keeper.InitGenesis(s.ctx, tc.genesis)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestExportGenesis() {
	testCases := []struct {
		name     string
		setup    func()
		expCount uint64
	}{
		{
			name: "export after init with zero",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
				s.Require().NoError(err)
			},
			expCount: 0,
		},
		{
			name: "export after init with value",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 42})
				s.Require().NoError(err)
			},
			expCount: 42,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			genesis, err := s.keeper.ExportGenesis(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expCount, genesis.Count)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
