package counter

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/example/x/counter/keeper"
	countertypes "github.com/cosmos/example/x/counter/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

const (
	ModuleName = "counter"
	StoreKey   = ModuleName
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}

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
	return ModuleName
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
