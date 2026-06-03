package dex

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/example/x/dex/keeper"
	dextypes "github.com/cosmos/example/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func (a AppModuleBasic) Name() string { return dextypes.ModuleName }

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := dextypes.GenesisState{
		Pools: []dextypes.Pool{
			{
				Id:       "1",
				DenomA:   "stake",
				DenomB:   "tokenb",
				ReserveA: "1000000000000000000",
				ReserveB: "1000000000000000000",
			},
		},
	}
	return cdc.MustMarshalJSON(&gs)
}

func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

func (a AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
	dextypes.RegisterInterfaces(registry)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := dextypes.RegisterQueryHandlerClient(clientCtx.CmdContext, mux, dextypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	amino.RegisterConcrete(&dextypes.GenesisState{}, "example/DexGenesisState", nil)
}

type AppModule struct {
	keeper *keeper.Keeper
	AppModuleBasic
}

func (a AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

func (a AppModule) EndBlock(ctx context.Context) error {
	return nil
}

func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper:         keeper,
		AppModuleBasic: AppModuleBasic{cdc: cdc},
	}
}

func (a AppModule) RegisterServices(configurator module.Configurator) {
	dextypes.RegisterMsgServer(configurator.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
	dextypes.RegisterQueryServer(configurator.QueryServer(), keeper.NewQueryServer(a.keeper))
}

func (a AppModule) InitGenesis(ctx sdk.Context, jsonCodec codec.JSONCodec, message json.RawMessage) {
	gs := &dextypes.GenesisState{}
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

func (a AppModule) HasServices() bool {
	return true
}

func (a AppModule) HasGenesis() bool {
	return true
}

func (a AppModule) IsAppModule() {
}

func (a AppModule) IsOnePerModuleType() {
}
