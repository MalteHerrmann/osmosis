package app

import (
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	evmosanteinterfaces "github.com/evmos/os/ante/interfaces"

	osmoante "github.com/osmosis-labs/osmosis/v26/ante"
	v9 "github.com/osmosis-labs/osmosis/v26/app/upgrades/v9"

	corestoretypes "cosmossdk.io/core/store"

	smartaccountante "github.com/osmosis-labs/osmosis/v26/x/smart-account/ante"
	smartaccountkeeper "github.com/osmosis-labs/osmosis/v26/x/smart-account/keeper"

	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"

	txfeeskeeper "github.com/osmosis-labs/osmosis/v26/x/txfees/keeper"
	txfeestypes "github.com/osmosis-labs/osmosis/v26/x/txfees/types"

	auctionante "github.com/skip-mev/block-sdk/v2/x/auction/ante"

	evmoscosmosante "github.com/evmos/os/ante/cosmos"
	evmosevmante "github.com/evmos/os/ante/evm"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
	evmtypes "github.com/evmos/os/x/evm/types"
	feemarketkeeper "github.com/evmos/os/x/feemarket/keeper"
)

// BlockSDKAnteHandlerParams are the parameters necessary to configure the block-sdk antehandlers
type BlockSDKAnteHandlerParams struct {
	mevLane       auctionante.MEVLane
	auctionKeeper auctionkeeper.Keeper
	txConfig      client.TxConfig
}

type HandlerOptions struct {
	appOpts             servertypes.AppOptions
	wasmConfig          wasmtypes.WasmConfig
	txCounterStoreKey   corestoretypes.KVStoreService
	accountKeeper       evmosanteinterfaces.AccountKeeper
	smartAccountKeeper  *smartaccountkeeper.Keeper
	bankKeeper          txfeestypes.BankKeeper
	txFeesKeeper        *txfeeskeeper.Keeper
	spotPriceCalculator txfeestypes.SpotPriceCalculator
	sigGasConsumer      ante.SignatureVerificationGasConsumer
	signModeHandler     *txsigning.HandlerMap
	channelKeeper       *ibckeeper.Keeper
	blockSDKParams      BlockSDKAnteHandlerParams
	appCodec            codec.Codec
	evmKeeper           *evmkeeper.Keeper
	feemarketKeeper     feemarketkeeper.Keeper
	maxGasWanted        uint64
}

func (options HandlerOptions) Validate() error {
	if options.appOpts == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "appOpts cannot be nil")
	}

	if options.txCounterStoreKey == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "txCounterStoreKey cannot be nil")
	}

	if options.blockSDKParams.txConfig == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "txConfig cannot be nil")
	}

	if options.blockSDKParams.mevLane == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "mevLane cannot be nil")
	}

	if options.smartAccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "smartAccountKeeper cannot be nil")
	}

	if options.txFeesKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "txFeesKeeper cannot be nil")
	}

	if options.sigGasConsumer == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sigGasConsumer cannot be nil")
	}

	if options.signModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "signModeHandler cannot be nil")
	}

	if options.channelKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "channelKeeper cannot be nil")
	}

	if options.evmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evmKeeper cannot be nil")
	}

	return nil
}

func NewHandlerOptions(
	appOpts servertypes.AppOptions,
	wasmConfig wasmtypes.WasmConfig,
	txCounterStoreKey corestoretypes.KVStoreService,
	accountKeeper evmosanteinterfaces.AccountKeeper,
	smartAccountKeeper *smartaccountkeeper.Keeper,
	bankKeeper txfeestypes.BankKeeper,
	txFeesKeeper *txfeeskeeper.Keeper,
	spotPriceCalculator txfeestypes.SpotPriceCalculator,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler *txsigning.HandlerMap,
	channelKeeper *ibckeeper.Keeper,
	blockSDKParams BlockSDKAnteHandlerParams,
	appCodec codec.Codec,
	evmKeeper *evmkeeper.Keeper,
	feemarketKeeper feemarketkeeper.Keeper,
	maxGasWanted uint64,
) HandlerOptions {
	return HandlerOptions{
		appOpts:             appOpts,
		wasmConfig:          wasmConfig,
		txCounterStoreKey:   txCounterStoreKey,
		accountKeeper:       accountKeeper,
		smartAccountKeeper:  smartAccountKeeper,
		bankKeeper:          bankKeeper,
		txFeesKeeper:        txFeesKeeper,
		spotPriceCalculator: spotPriceCalculator,
		sigGasConsumer:      sigGasConsumer,
		signModeHandler:     signModeHandler,
		channelKeeper:       channelKeeper,
		blockSDKParams:      blockSDKParams,
		appCodec:            appCodec,
		evmKeeper:           evmKeeper,
		feemarketKeeper:     feemarketKeeper,
		maxGasWanted:        maxGasWanted,
	}
}

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(options HandlerOptions) sdk.AnteHandler {
	// TODO: check if this is fine, prior the mempool decorator was always instantiated, which was setting the
	// backup file path.
	// Now, since the decorators are only instantiated for non-EVM transactions, we have to manually call this here.
	options.txFeesKeeper.SetBackupFilePath()

	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		txWithExtensions, ok := tx.(ante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/os.evm.v1.ExtensionOptionsEthereumTx":
					// handle as *evmtypes.MsgEthereumTx
					anteHandler = newMonoEVMAnteHandler(options)
				case "/os.types.v1.ExtensionOptionDynamicFeeTx":
					// cosmos-sdk tx with dynamic fee extension
					anteHandler = NewCosmosAnteHandler(options)
				// TODO: check with Paddy which extensions are supported on Osmosis
				default:
					return ctx, errorsmod.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// handle as normal Cosmos SDK tx
		switch tx.(type) {
		case sdk.Tx:
			anteHandler = NewCosmosAnteHandler(options)
		default:
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}
}

// NewCosmosAnteHandler creates the decorator chain used for Cosmos transactions.
//
// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
func NewCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	mempoolFeeOptions := txfeestypes.NewMempoolFeeOptions(options.appOpts)
	mempoolFeeDecorator := txfeeskeeper.NewMempoolFeeDecorator(*options.txFeesKeeper, mempoolFeeOptions)
	sendblockOptions := osmoante.NewSendBlockOptions(options.appOpts)
	sendblockDecorator := osmoante.NewSendBlockDecorator(sendblockOptions, options.appCodec)
	deductFeeDecorator := txfeeskeeper.NewDeductFeeDecorator(*options.txFeesKeeper, options.accountKeeper, options.bankKeeper, nil)

	// classicSignatureVerificationDecorator is the old flow to enable a circuit breaker
	classicSignatureVerificationDecorator := sdk.ChainAnteDecorators(
		deductFeeDecorator,
		// We use the old pubkey decorator here to ensure that accounts work as expected,
		// in SetPubkeyDecorator we set a pubkey in the account store, for authenticators
		// we avoid this code path completely.
		ante.NewSetPubKeyDecorator(options.accountKeeper),
		ante.NewValidateSigCountDecorator(options.accountKeeper),
		ante.NewSigGasConsumeDecorator(options.accountKeeper, options.sigGasConsumer),
		ante.NewSigVerificationDecorator(options.accountKeeper, options.signModeHandler),
		ante.NewIncrementSequenceDecorator(options.accountKeeper),
		ibcante.NewRedundantRelayDecorator(options.channelKeeper),
		// auction module antehandler
		auctionante.NewAuctionDecorator(
			options.blockSDKParams.auctionKeeper,
			options.blockSDKParams.txConfig.TxEncoder(),
			options.blockSDKParams.mevLane,
		),
	)

	// authenticatorVerificationDecorator is the new authenticator flow that's embedded into the circuit breaker ante
	authenticatorVerificationDecorator := sdk.ChainAnteDecorators(
		smartaccountante.NewEmitPubKeyDecoratorEvents(options.accountKeeper),
		ante.NewValidateSigCountDecorator(options.accountKeeper), // we can probably remove this as multisigs are not supported here
		// Both the signature verification, fee deduction, and gas consumption functionality
		// is embedded in the authenticator decorator
		smartaccountante.NewAuthenticatorDecorator(options.appCodec, options.smartAccountKeeper, options.accountKeeper, options.signModeHandler, deductFeeDecorator),
		ante.NewIncrementSequenceDecorator(options.accountKeeper),
		// auction module antehandler
		auctionante.NewAuctionDecorator(
			options.blockSDKParams.auctionKeeper,
			options.blockSDKParams.txConfig.TxEncoder(),
			options.blockSDKParams.mevLane,
		),
	)

	return sdk.ChainAnteDecorators(
		evmoscosmosante.NewRejectMessagesDecorator(), // reject all EVM transactions
		evmoscosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
		),
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(options.wasmConfig.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.txCounterStoreKey),
		// // TODO: should this be removed? We are rejecting all unknown extension options currently in the main ante handler definition.
		// // Also this would be rejecting dynamic fee transactions currently, which are needed for EIP-1559 transactions to work.
		// ante.NewExtensionOptionsDecorator(nil),
		v9.MsgFilterDecorator{},
		// Use Mempool Fee Decorator from our txfees module instead of default one from auth
		// https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/middleware/fee.go#L34
		mempoolFeeDecorator,
		sendblockDecorator,
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(options.accountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.accountKeeper),
		smartaccountante.NewCircuitBreakerDecorator(
			options.smartAccountKeeper,
			authenticatorVerificationDecorator,
			classicSignatureVerificationDecorator,
		),
	)
}

// newMonoEVMAnteHandler is the ante handler for EVM transactions.
func newMonoEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	monoEVMDecorator := evmosevmante.NewEVMMonoDecorator(
		options.accountKeeper,
		options.feemarketKeeper,
		options.evmKeeper,
		options.maxGasWanted,
	)

	return sdk.ChainAnteDecorators(
		monoEVMDecorator,
	)
}
