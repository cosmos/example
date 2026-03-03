package keeper_test

import (
	"context"
	"errors"
	"math"
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

// MockBankKeeper is a mock implementation of the BankKeeper interface
type MockBankKeeper struct {
	SendCoinsFromAccountToModuleFn func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.SendCoinsFromAccountToModuleFn != nil {
		return m.SendCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
	}
	return nil
}

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	keeper      *keeper.Keeper
	queryClient types.QueryClient
	msgServer   types.MsgServer
	bankKeeper  *MockBankKeeper
	authority   string
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey("counter")
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	s.authority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	s.bankKeeper = &MockBankKeeper{}
	k := keeper.NewKeeper(storeService, encCfg.Codec, s.bankKeeper, keeper.WithAuthority(s.authority))

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
		{
			name: "with params",
			genesis: &types.GenesisState{
				Count: 50,
				Params: types.Params{
					MaxAddValue: 1000,
					AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
				},
			},
			expErr: false,
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
		name      string
		setup     func()
		expCount  uint64
		expParams types.Params
	}{
		{
			name: "export after init with zero",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
				s.Require().NoError(err)
			},
			expCount:  0,
			expParams: types.Params{},
		},
		{
			name: "export after init with value",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 42})
				s.Require().NoError(err)
			},
			expCount:  42,
			expParams: types.Params{},
		},
		{
			name: "export with params",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count: 100,
					Params: types.Params{
						MaxAddValue: 500,
						AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
					},
				})
				s.Require().NoError(err)
			},
			expCount: 100,
			expParams: types.Params{
				MaxAddValue: 500,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			genesis, err := s.keeper.ExportGenesis(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expCount, genesis.Count)
			s.Require().Equal(tc.expParams, genesis.Params)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestGetCount() {
	testCases := []struct {
		name     string
		setup    func()
		expCount uint64
		expErr   bool
	}{
		{
			name: "get count from uninitialized state",
			setup: func() {
				// Don't initialize genesis - counter should return 0
			},
			expCount: 0,
			expErr:   false,
		},
		{
			name: "get count after init with zero",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
				s.Require().NoError(err)
			},
			expCount: 0,
			expErr:   false,
		},
		{
			name: "get count after init with value",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 42})
				s.Require().NoError(err)
			},
			expCount: 42,
			expErr:   false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			count, err := s.keeper.GetCount(s.ctx)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expCount, count)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAddCount() {
	testCases := []struct {
		name         string
		setup        func()
		sender       string
		amount       uint64
		expErr       bool
		expErrMsg    string
		expPostCount uint64
	}{
		{
			name: "add to zero counter",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  0,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
			},
			sender:       "cosmos1test",
			amount:       10,
			expErr:       false,
			expPostCount: 10,
		},
		{
			name: "add to existing counter",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  50,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
			},
			sender:       "cosmos1test",
			amount:       25,
			expErr:       false,
			expPostCount: 75,
		},
		{
			name: "add zero",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  100,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
			},
			sender:       "cosmos1test",
			amount:       0,
			expErr:       false,
			expPostCount: 100,
		},
		{
			name: "add max - should error",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  100,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
			},
			sender:       "cosmos1test",
			amount:       math.MaxUint64,
			expErr:       true,
			expPostCount: 100,
		},
		{
			name: "add exceeds max_add_value - should error",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  0,
					Params: types.Params{MaxAddValue: 50},
				})
				s.Require().NoError(err)
			},
			sender:    "cosmos1test",
			amount:    100,
			expErr:    true,
			expErrMsg: "exceeds max allowed",
		},
		{
			name: "add with cost - charges user",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count: 0,
					Params: types.Params{
						MaxAddValue: 100,
						AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
					},
				})
				s.Require().NoError(err)
				s.bankKeeper.SendCoinsFromAccountToModuleFn = func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
					s.Require().Equal(types.ModuleName, recipientModule)
					s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("stake", 10)), amt)
					return nil
				}
			},
			sender:       "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			amount:       10,
			expErr:       false,
			expPostCount: 10,
		},
		{
			name: "add with cost - insufficient funds",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count: 0,
					Params: types.Params{
						MaxAddValue: 100,
						AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
					},
				})
				s.Require().NoError(err)
				s.bankKeeper.SendCoinsFromAccountToModuleFn = func(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
					return errors.New("insufficient funds")
				}
			},
			sender:    "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			amount:    10,
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			newCount, err := s.keeper.AddCount(s.ctx, tc.sender, tc.amount)
			if tc.expErr {
				s.Require().Error(err)
				if tc.expErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expErrMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expPostCount, newCount)

				// Verify the count is persisted
				count, err := s.keeper.GetCount(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.expPostCount, count)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetParams() {
	testCases := []struct {
		name      string
		setup     func()
		params    types.Params
		expErr    bool
		expParams types.Params
	}{
		{
			name: "set params with max add value",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{})
				s.Require().NoError(err)
			},
			params: types.Params{
				MaxAddValue: 500,
			},
			expErr: false,
			expParams: types.Params{
				MaxAddValue: 500,
			},
		},
		{
			name: "set params with add cost",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{})
				s.Require().NoError(err)
			},
			params: types.Params{
				MaxAddValue: 1000,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
			},
			expErr: false,
			expParams: types.Params{
				MaxAddValue: 1000,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
			},
		},
		{
			name: "update existing params",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Params: types.Params{
						MaxAddValue: 100,
						AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 10)),
					},
				})
				s.Require().NoError(err)
			},
			params: types.Params{
				MaxAddValue: 200,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 20)),
			},
			expErr: false,
			expParams: types.Params{
				MaxAddValue: 200,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 20)),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			err := s.keeper.SetParams(s.ctx, tc.params)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// Verify params were set
				params, err := s.keeper.GetParams(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.expParams.MaxAddValue, params.MaxAddValue)
				s.Require().Equal(tc.expParams.AddCost, params.AddCost)
			}
		})
	}
}
