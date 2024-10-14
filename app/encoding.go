package app

import (
	"github.com/osmosis-labs/osmosis/v26/app/keepers"
	"github.com/osmosis-labs/osmosis/v26/app/params"

	evmosenccodec "github.com/evmos/os/encoding/codec"
)

var encodingConfig params.EncodingConfig = MakeEncodingConfig()

func GetEncodingConfig() params.EncodingConfig {
	return encodingConfig
}

// MakeEncodingConfig creates an EncodingConfig.
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()

	// NOTE: the evmOS functions also register the standard Cosmos SDK interfaces and codecs
	evmosenccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	evmosenccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	keepers.AppModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}
