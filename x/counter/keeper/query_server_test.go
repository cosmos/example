package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/example/x/counter/types"
)

func (s *KeeperTestSuite) TestQueryCount() {
	testCases := []struct {
		name     string
		setup    func()
		req      *types.QueryCountRequest
		expErr   bool
		expCount uint64
	}{
		{
			name: "query zero count",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
				s.Require().NoError(err)
			},
			req:      &types.QueryCountRequest{},
			expErr:   false,
			expCount: 0,
		},
		{
			name: "query non-zero count",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 999})
				s.Require().NoError(err)
			},
			req:      &types.QueryCountRequest{},
			expErr:   false,
			expCount: 999,
		},
		{
			name: "query after add",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  10,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
				_, err = s.msgServer.Add(s.ctx, &types.MsgAddRequest{Sender: "cosmos1test", Add: 5})
				s.Require().NoError(err)
			},
			req:      &types.QueryCountRequest{},
			expErr:   false,
			expCount: 15,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			resp, err := s.queryClient.Count(s.ctx, tc.req)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().Equal(tc.expCount, resp.Count)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryParams() {
	testCases := []struct {
		name      string
		setup     func()
		req       *types.QueryParamsRequest
		expErr    bool
		expParams types.Params
	}{
		{
			name: "query default params",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{})
				s.Require().NoError(err)
			},
			req:       &types.QueryParamsRequest{},
			expErr:    false,
			expParams: types.Params{},
		},
		{
			name: "query custom params",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Params: types.Params{
						MaxAddValue: 500,
						AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
					},
				})
				s.Require().NoError(err)
			},
			req:    &types.QueryParamsRequest{},
			expErr: false,
			expParams: types.Params{
				MaxAddValue: 500,
				AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			resp, err := s.queryClient.Params(s.ctx, tc.req)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().NotNil(resp.Params)
				s.Require().Equal(tc.expParams.MaxAddValue, resp.Params.MaxAddValue)
				s.Require().Equal(tc.expParams.AddCost.String(), resp.Params.AddCost.String())
			}
		})
	}
}
