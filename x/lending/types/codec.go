package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	// TODO: Register the modules msgs
	cdc.RegisterConcrete(MsgCreateDebt{}, "lending/CreateDebt", nil)
	cdc.RegisterConcrete(MsgPayDebt{}, "lending/PayDebt", nil)
	cdc.RegisterConcrete(MsgChangeDebt{}, "lending/ChangeDebt", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
