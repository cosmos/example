package counter

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/example/x/counter/keeper"
	"github.com/cosmos/example/x/counter/simulation"
	countertypes "github.com/cosmos/example/x/counter/types"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.HasGenesisBasics    = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}

	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func (a AppModuleBasic) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	gs := countertypes.GenesisState{}
	return jsonCodec.MustMarshalJSON(&gs)
}

func (a AppModuleBasic) ValidateGenesis(jsonCodec codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	gs := countertypes.GenesisState{}
	return jsonCodec.UnmarshalJSON(message, &gs)
}

func (a AppModuleBasic) Name() string {
	return countertypes.ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {}

func (a AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
	countertypes.RegisterInterfaces(registry)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := countertypes.RegisterQueryHandlerClient(clientCtx.CmdContext, mux, countertypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

type AppModule struct {
	keeper *keeper.Keeper
	AppModuleBasic
}

func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper:         keeper,
		AppModuleBasic: AppModuleBasic{cdc: cdc},
	}
}

func (a AppModule) RegisterServices(configurator module.Configurator) {
	countertypes.RegisterMsgServer(configurator.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
	countertypes.RegisterQueryServer(configurator.QueryServer(), keeper.NewQueryServer(a.keeper))
}

func (a AppModule) DefaultGenesis(jsonCodec codec.JSONCodec) json.RawMessage {
	gs := countertypes.GenesisState{Count: 0}
	return jsonCodec.MustMarshalJSON(&gs)

}

func (a AppModule) ValidateGenesis(jsonCodec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	gs := &countertypes.GenesisState{}
	if err := jsonCodec.UnmarshalJSON(message, gs); err != nil {
		return err
	}
	return nil
}

func (a AppModule) InitGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {
	gs := &countertypes.GenesisState{}
	jsonCodec.MustUnmarshalJSON(message, gs)
	if err := a.keeper.InitGenesis(ctx, gs); err != nil {
		panic(err)
	}
}

func (a AppModule) ExportGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	gs, err := a.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return jsonCodec.MustMarshalJSON(gs)
}

func (a AppModule) ConsensusVersion() uint64 { return 1 }

func (a AppModule) IsOnePerModuleType() {}

func (a AppModule) IsAppModule() {}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the counter module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for counter module's types.
func (a AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[countertypes.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(a.keeper.Schema)
}

// WeightedOperations returns nil - use WeightedOperationsX instead.
func (a AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// WeightedOperationsX registers weighted counter module operations for simulation.
func (a AppModule) WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry) {
	reg.Add(weights.Get("msg_add", 100), simulation.MsgAddFactory())
}
