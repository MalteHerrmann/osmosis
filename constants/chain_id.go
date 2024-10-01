package constants

import "fmt"

const (
	ChainIDPrefix = "osmosis"
	EIP155ChainID = 9009
	ChainIDSuffix = 1
)

var (
	MainnetChainID = fmt.Sprintf("%s_%d-%d", ChainIDPrefix, EIP155ChainID, ChainIDSuffix)
)
