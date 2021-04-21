package common

import (
	wrappers "github.com/DigiU-Lab/eth-contracts-go-wrappers"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetContractByAddressNetwork1(config config.AppConfig, ethClient *ethclient.Client) (out *wrappers.Bridge, err error) {
	return wrappers.NewBridge(common.HexToAddress(config.BRIDGE_ADDRESS_1), ethClient)
}
