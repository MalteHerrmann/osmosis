package wasmbinding

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app"
	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	osmoconstants "github.com/osmosis-labs/osmosis/v25/constants"
)

func CreateTestInput() (*app.OsmosisApp, sdk.Context) {
	osmosis := app.Setup(false)
	ctx := osmosis.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1, ChainID: osmoconstants.MainnetChainID, Time: time.Now().UTC()})
	return osmosis, ctx
}

func FundAccount(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, acct sdk.AccAddress) {
	t.Helper()
	err := testutil.FundAccount(ctx, osmosis.BankKeeper, acct, sdk.NewCoins(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000000)),
	))
	require.NoError(t, err)
}

// we need to make this deterministic (same every test run), as content might affect gas costs
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}
