package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmosante "github.com/evmos/os/ante"
	"github.com/evmos/os/crypto/ethsecp256k1"
)

var _ authante.SignatureVerificationGasConsumer = ExtendedSigVerificationGasConsumer

// ExtendedSigVerificationGasConsumer is based on the default implementation of SignatureVerificationGasConsumer
// from the Cosmos SDK and is extended to support the eth_secp256k1 public key type.
//
// It consumes gas for signature verification based upon the public key type.
// The cost is fetched from the given params and is matched by the concrete type.
//
// TODO: Is this desired? This would add support for signing Cosmos transactions with eth_secp256k1 keys - not only EVM transactions.
func ExtendedSigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {
	case *ed25519.PubKey:
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "ED25519 public keys are unsupported")

	case *secp256k1.PubKey:
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return nil

	case *secp256r1.PubKey:
		meter.ConsumeGas(params.SigVerifyCostSecp256r1(), "ante verify: secp256r1")
		return nil

	case multisig.PubKey:
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}
		err := authante.ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)
		if err != nil {
			return err
		}
		return nil

	// evmOS key support
	case *ethsecp256k1.PubKey:
		meter.ConsumeGas(evmosante.Secp256k1VerifyCost, "ante verify: ethsecp256k1")
		return nil

	default:
		return errorsmod.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}
