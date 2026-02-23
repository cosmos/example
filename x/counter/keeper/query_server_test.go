package keeper_test

import (
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
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 10})
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

func (s *KeeperTestSuite) TestQueryCountDirectKeeper() {
	s.SetupTest()
	err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 123})
	s.Require().NoError(err)

	queryServer := s.keeper
	_ = queryServer

	resp, err := s.queryClient.Count(s.ctx, &types.QueryCountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(123), resp.Count)
}
