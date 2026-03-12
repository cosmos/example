package keeper_test

import (
	"context"
	"errors"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
					Count:  0,
					Params: types.Params{MaxAddValue: 100},
				})
				s.Require().NoError(err)
			},
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 10},
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
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 25},
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
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: 0},
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
			msg:          &types.MsgAddRequest{Sender: "cosmos1test", Add: math.MaxUint64},
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
			msg:       &types.MsgAddRequest{Sender: "cosmos1test", Add: 100},
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
			msg:          &types.MsgAddRequest{Sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", Add: 10},
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
			msg:       &types.MsgAddRequest{Sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", Add: 10},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			resp, err := s.msgServer.Add(s.ctx, tc.msg)
			if tc.expErr {
				s.Require().Error(err)
				if tc.expErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expErrMsg)
				}
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
	err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{
		Count:  0,
		Params: types.Params{MaxAddValue: 100},
	})
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

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	testCases := []struct {
		name      string
		setup     func()
		msg       *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid update params by authority",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{})
				s.Require().NoError(err)
			},
			msg: &types.MsgUpdateParams{
				Authority: s.authority,
				Params: types.Params{
					MaxAddValue: 500,
					AddCost:     sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
				},
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			setup: func() {
				err := s.keeper.InitGenesis(s.ctx, &types.GenesisState{})
				s.Require().NoError(err)
			},
			msg: &types.MsgUpdateParams{
				Authority: "cosmos1invalid",
				Params: types.Params{
					MaxAddValue: 500,
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			resp, err := s.msgServer.UpdateParams(s.ctx, tc.msg)
			if tc.expErr {
				s.Require().Error(err)
				if tc.expErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expErrMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)

				// Verify params were updated
				params, err := s.keeper.GetParams(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.msg.Params.MaxAddValue, params.MaxAddValue)
				s.Require().Equal(tc.msg.Params.AddCost, params.AddCost)
			}
		})
	}
}
