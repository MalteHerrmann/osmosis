package v21

import (
	"fmt"

	store "cosmossdk.io/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/osmosis-labs/osmosis/v25/app/upgrades"
	osmoconstants "github.com/osmosis-labs/osmosis/v25/constants"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v21 upgrade.
const (
	UpgradeName = "v21"
)

var (
	TestingChainId = fmt.Sprintf("testingchain_%d-%d", osmoconstants.EIP155ChainID, osmoconstants.ChainIDSuffix)
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// v47 modules
			crisistypes.ModuleName,
			consensustypes.ModuleName,
		},
		Deleted: []string{},
	},
}
