package tests

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/example"
	"github.com/cosmos/example/x/counter/keeper"
	countertypes "github.com/cosmos/example/x/counter/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CounterIntegrationTestSuite struct {
	suite.Suite
	app *example.ExampleApp
	ctx sdk.Context
}

func (s *CounterIntegrationTestSuite) SetupTest() {
	s.app = Setup(s.T())
	s.ctx = s.app.NewContext(false)
}

func (s *CounterIntegrationTestSuite) TestCounterInitialState() {
	resp, err := s.app.CounterKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterAddViaKeeper() {
	err := s.app.CounterKeeper.InitGenesis(s.ctx, &countertypes.GenesisState{Count: 10})
	s.Require().NoError(err)

	resp, err := s.app.CounterKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(10), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterQueryServer() {
	err := s.app.CounterKeeper.InitGenesis(s.ctx, &countertypes.GenesisState{Count: 42})
	s.Require().NoError(err)

	queryServer := keeper.NewQueryServer(s.app.CounterKeeper)
	resp, err := queryServer.Count(s.ctx, &countertypes.QueryCountRequest{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(42), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterMsgServiceDirect() {
	err := s.app.CounterKeeper.InitGenesis(s.ctx, &countertypes.GenesisState{Count: 0})
	s.Require().NoError(err)

	msgServer := s.app.MsgServiceRouter().Handler(&countertypes.MsgAddRequest{})
	s.Require().NotNil(msgServer)

	result, err := msgServer(s.ctx, &countertypes.MsgAddRequest{
		Sender: "cosmos1test",
		Add:    100,
	})
	s.Require().NoError(err)
	s.Require().NotNil(result)

	resp, err := s.app.CounterKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(100), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterMultipleAdds() {
	err := s.app.CounterKeeper.InitGenesis(s.ctx, &countertypes.GenesisState{Count: 0})
	s.Require().NoError(err)

	msgServer := s.app.MsgServiceRouter().Handler(&countertypes.MsgAddRequest{})

	for i := 0; i < 5; i++ {
		_, err := msgServer(s.ctx, &countertypes.MsgAddRequest{
			Sender: "cosmos1test",
			Add:    10,
		})
		s.Require().NoError(err)
	}

	resp, err := s.app.CounterKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(50), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterWithCustomGenesis() {
	app := SetupWithCustomGenesis(s.T(), func(gs example.GenesisState) example.GenesisState {
		counterGenesis := countertypes.GenesisState{Count: 1000}
		gs[countertypes.ModuleName] = s.app.AppCodec().MustMarshalJSON(&counterGenesis)
		return gs
	})

	ctx := app.NewContext(false)
	resp, err := app.CounterKeeper.ExportGenesis(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1000), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterStatePersistedAcrossBlocks() {
	msgServer := keeper.NewMsgServerImpl(s.app.CounterKeeper)

	_, err := msgServer.Add(s.ctx, &countertypes.MsgAddRequest{
		Sender: "cosmos1test",
		Add:    25,
	})
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.app.LastBlockHeight() + 1,
		Hash:   s.app.LastCommitID().Hash,
	})
	s.Require().NoError(err)

	_, err = s.app.Commit()
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.app.LastBlockHeight() + 1,
		Hash:   s.app.LastCommitID().Hash,
	})
	s.Require().NoError(err)

	newCtx := s.app.NewContext(false)
	resp, err := s.app.CounterKeeper.ExportGenesis(newCtx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(25), resp.Count)

	_, err = msgServer.Add(newCtx, &countertypes.MsgAddRequest{
		Sender: "cosmos1test",
		Add:    15,
	})
	s.Require().NoError(err)

	resp, err = s.app.CounterKeeper.ExportGenesis(newCtx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(40), resp.Count)
}

func (s *CounterIntegrationTestSuite) TestCounterGenesisExportImport() {
	err := s.app.CounterKeeper.InitGenesis(s.ctx, &countertypes.GenesisState{Count: 999})
	s.Require().NoError(err)

	exported, err := s.app.CounterKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)

	exportedBytes, err := json.Marshal(exported)
	s.Require().NoError(err)

	var imported countertypes.GenesisState
	err = json.Unmarshal(exportedBytes, &imported)
	s.Require().NoError(err)

	s.Require().Equal(uint64(999), imported.Count)
}

func TestCounterIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CounterIntegrationTestSuite))
}
