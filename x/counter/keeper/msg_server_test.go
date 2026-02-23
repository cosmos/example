package keeper_test

import (
	"github.com/cosmos/example/x/counter/types"
)

func (s *KeeperTestSuite) TestMsgAdd() {
	testCases := []struct {
		name         string
		setup        func()
		msg          *types.MsgAddRequest
		expErr       bool
		expErrMsg    string
		expPostCount uint64
	}{
		{
			name: "add to zero counter",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
				s.Require().NoError(err)
			},
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 10},
			expErr:       false,
			expPostCount: 10,
		},
		{
			name: "add to existing counter",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 50})
				s.Require().NoError(err)
			},
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 25},
			expErr:       false,
			expPostCount: 75,
		},
		{
			name: "add zero",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 100})
				s.Require().NoError(err)
			},
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 0},
			expErr:       false,
			expPostCount: 100,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			resp, err := s.msgServer.Add(s.ctx, tc.msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().Equal(tc.expPostCount, resp.UpdatedCount)

				queryResp, err := s.queryClient.Count(s.ctx, &types.QueryCountRequest{})
				s.Require().NoError(err)
				s.Require().Equal(tc.expPostCount, queryResp.Count)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgAddEmitsEvent() {
	s.SetupTest()
	err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{Count: 0})
	s.Require().NoError(err)

	_, err = s.msgServer.Add(s.ctx, &types.MsgAddRequest{Sender: "cosmos1test", Add: 42})
	s.Require().NoError(err)

	events := s.ctx.EventManager().Events()
	s.Require().NotEmpty(events)

	found := false
	for _, event := range events {
		if event.Type == "count_increased" {
			found = true
			for _, attr := range event.Attributes {
				if attr.Key == "count" {
					s.Require().Equal("42", attr.Value)
				}
			}
		}
	}
	s.Require().True(found, "count_increased event not found")
}
