package cmd

import (
	cosmoshd "github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	evmoshd "github.com/evmos/os/crypto/hd"
)

// ExtendedKeyringOption returns a keyring.Option that adds support for the eth_secp256k1 curve
// in addition to the default Cosmos secp256k1 curve.
func ExtendedKeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = keyring.SigningAlgoList{cosmoshd.Secp256k1, evmoshd.EthSecp256k1}
	}
}
