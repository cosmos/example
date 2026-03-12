package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"

	countertypes "github.com/cosmos/example/x/counter/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	conn    *grpc.ClientConn
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	var err error
	s.cfg = network.DefaultConfig(NewTestNetworkFixture)
	s.cfg.NumValidators = 1

	// Customize counter genesis to set initial count and permissive params
	genesisState := s.cfg.GenesisState
	counterGenesis := countertypes.GenesisState{
		Count: 0,
		Params: countertypes.Params{
			MaxAddValue: 1000,
			AddCost:     nil, // No cost for testing
		},
	}
	counterGenesisBz, err := s.cfg.Codec.MarshalJSON(&counterGenesis)
	s.Require().NoError(err)
	genesisState[countertypes.ModuleName] = counterGenesisBz
	s.cfg.GenesisState = genesisState

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	val0 := s.network.Validators[0]
	s.conn, err = grpc.NewClient(
		val0.AppConfig.GRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.cfg.InterfaceRegistry).GRPCCodec())),
	)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.conn.Close()
	s.network.Cleanup()
}

// getCurrentCount is a helper to get the current counter value
func (s *E2ETestSuite) getCurrentCount() uint64 {
	queryClient := countertypes.NewQueryClient(s.conn)
	resp, err := queryClient.Count(context.Background(), &countertypes.QueryCountRequest{})
	s.Require().NoError(err)
	return resp.Count
}

// TestQueryCount tests querying the counter via gRPC
func (s *E2ETestSuite) TestQueryCount() {
	queryClient := countertypes.NewQueryClient(s.conn)

	// Just verify we can query - the value depends on test execution order
	resp, err := queryClient.Count(context.Background(), &countertypes.QueryCountRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}

// TestQueryParams tests querying the module params via gRPC
func (s *E2ETestSuite) TestQueryParams() {
	queryClient := countertypes.NewQueryClient(s.conn)

	resp, err := queryClient.Params(context.Background(), &countertypes.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Params)
	// Verify the params match what was set in genesis
	s.Require().Equal(uint64(1000), resp.Params.MaxAddValue)
}

// TestAddCounter tests broadcasting an Add transaction via gRPC
func (s *E2ETestSuite) TestAddCounter() {
	val := s.network.Validators[0]

	// Query initial count
	initialCount := s.getCurrentCount()

	// Build and broadcast Add transaction
	txBuilder := s.mkCounterAddTx(val, 42)
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	txClient := txtypes.NewServiceClient(s.conn)
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), grpcRes.TxResponse.Code, "tx failed: %s", grpcRes.TxResponse.RawLog)

	// Wait for tx to be included in a block
	s.Require().NoError(s.network.WaitForNextBlock())

	// Query updated count - should have increased by 42
	finalCount := s.getCurrentCount()
	s.Require().Equal(initialCount+42, finalCount)
}

// TestMultipleAdds tests broadcasting multiple Add transactions
func (s *E2ETestSuite) TestMultipleAdds() {
	val := s.network.Validators[0]

	// Query initial count
	initialCount := s.getCurrentCount()

	// Send multiple add transactions
	addValues := []uint64{10, 20, 30}
	for _, addValue := range addValues {
		txBuilder := s.mkCounterAddTx(val, addValue)
		txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
		s.Require().NoError(err)

		txClient := txtypes.NewServiceClient(s.conn)
		grpcRes, err := txClient.BroadcastTx(
			context.Background(),
			&txtypes.BroadcastTxRequest{
				Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
				TxBytes: txBytes,
			},
		)
		s.Require().NoError(err)
		s.Require().Equal(uint32(0), grpcRes.TxResponse.Code, "tx failed: %s", grpcRes.TxResponse.RawLog)

		// Wait for each tx
		s.Require().NoError(s.network.WaitForNextBlock())
	}

	// Query final count - should have increased by 60
	finalCount := s.getCurrentCount()
	s.Require().Equal(initialCount+10+20+30, finalCount)
}

// TestAddExceedsMaxValue tests that adding more than MaxAddValue fails
func (s *E2ETestSuite) TestAddExceedsMaxValue() {
	val := s.network.Validators[0]

	// Query initial count
	initialCount := s.getCurrentCount()

	// Try to add more than MaxAddValue (1000)
	txBuilder := s.mkCounterAddTx(val, 1001)
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	txClient := txtypes.NewServiceClient(s.conn)
	_, err = txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	s.Require().NoError(err)
	// Note: SYNC broadcast only checks if tx was accepted to mempool
	// The actual validation error occurs in DeliverTx

	// Wait for the block
	s.Require().NoError(s.network.WaitForNextBlock())

	// The count should NOT have increased (tx should have failed in DeliverTx)
	finalCount := s.getCurrentCount()
	s.Require().Equal(initialCount, finalCount, "counter should not have changed when tx fails validation")
}

// mkCounterAddTx creates a signed MsgAddRequest transaction
func (s *E2ETestSuite) mkCounterAddTx(val *network.Validator, addValue uint64) client.TxBuilder {
	s.Require().NoError(s.network.WaitForNextBlock())

	txBuilder := val.ClientCtx.TxConfig.NewTxBuilder()
	feeAmount := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)}
	gasLimit := uint64(200000)

	s.Require().NoError(
		txBuilder.SetMsgs(&countertypes.MsgAddRequest{
			Sender: val.Address.String(),
			Add:    addValue,
		}),
	)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	txFactory := clienttx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false, true)
	s.Require().NoError(err)

	return txBuilder
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
