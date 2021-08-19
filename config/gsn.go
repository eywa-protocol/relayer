package config

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"
	"gopkg.in/yaml.v3"
	_ "gopkg.in/yaml.v3"
)

// Gsn is global object that holds all Gsn configuration variables.
var Gsn GsnConfig

type GsnConfig struct {
	TickerInterval time.Duration `yaml:"ticker_interval"`
	Chains         []*GsnChain   `yaml:"chains"`
	BootstrapAddrs []string      `yaml:"bootstrap-addrs"`
}

type GsnChain struct {
	Id               uint `yaml:"id"`
	ChainId          *big.Int
	RpcUrls          []string `yaml:"rpc_urls"`
	EcdsaKeyString   string   `yaml:"ecdsa_key"`
	EcdsaKey         *ecdsa.PrivateKey
	EcdsaAddress     common.Address
	ForwarderAddress common.Address `yaml:"forwarder_address"`
}

func LoadGsnConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file [%s] error: %w",
			path, err)
	}
	if err := yaml.Unmarshal(data, &Gsn); err != nil {
		return fmt.Errorf("unmarshal config error: %w", err)
	}

	for _, chain := range Gsn.Chains {
		if chain.EcdsaKey, err = crypto.HexToECDSA(strings.TrimPrefix(chain.EcdsaKeyString, "0x")); err != nil {
			return fmt.Errorf("decode chain [%d] ecdsa_key error: %w", chain.Id, err)
		}
		publicKey := chain.EcdsaKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("chain [%d] casting public key to ECDSA Address error: %w", chain.Id, err)
		}
		chain.EcdsaAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	}

	return nil
}

func (c *GsnChain) GetEthClient(skipUrl string) (client *ethclient.Client, url string, err error) {
	for _, url := range c.RpcUrls {
		if skipUrl != "" && len(c.RpcUrls) > 1 && url == skipUrl {
			continue
		} else if client, err = ethclient.Dial(url); err != nil {
			logrus.WithFields(logrus.Fields{
				field.CainId: c.Id,
				field.EthUrl: url,
			}).Error(fmt.Errorf("can not connect to chain rpc on error: %w", err))
		} else {

			return client, url, nil
		}
	}

	return nil, "", fmt.Errorf("connection to all rpc url for chain [%d]  failed", c.Id)
}
