package config

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
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

// Bridge is global object that holds all bridge configuration variables.
var Bridge Configuration

type Configuration struct {
	TickerInterval       time.Duration `yaml:"ticker_interval"`
	UptimeReportInterval time.Duration
	Rendezvous           string       `yaml:"rendezvous"`
	Chains               BridgeChains `yaml:"chains"`
	BootstrapAddrs       []string     `yaml:"bootstrap-addrs"`
	PromListenPort       *string      `yaml:"prom_listen_port"`
}

type BridgeChains []*BridgeChain

func (bcs BridgeChains) GetChainCfg(chainId uint) *BridgeChain {
	for _, bridgeChain := range bcs {
		if bridgeChain.Id == chainId {
			return bridgeChain
		}
	}
	return nil
}

type BridgeKeys struct {
}

type BridgeContract struct {
}

type BridgeChain struct {
	Id              uint `yaml:"id"`
	ChainId         *big.Int
	RpcUrls         []string `yaml:"rpc_urls"`
	EcdsaKeyString  string   `yaml:"ecdsa_key"`
	EcdsaKey        *ecdsa.PrivateKey
	EcdsaAddress    common.Address
	BridgeAddress   common.Address `yaml:"bridge_address"`
	NodeListAddress common.Address `yaml:"node_list_address"`
	DexPoolAddress  common.Address `yaml:"dex_pool_address"`
}

func LoadBridgeConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file [%s] error: %w",
			path, err)
	}
	if err := yaml.Unmarshal(data, &Bridge); err != nil {
		return fmt.Errorf("unmarshal config error: %w", err)
	}

	// todo: change to 24 hours
	Bridge.UptimeReportInterval = 5 * time.Minute

	for _, chain := range Bridge.Chains {
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

	// override config fields from env
	if rendezvous := os.Getenv("RANDEVOUE"); rendezvous != "" {
		Bridge.Rendezvous = rendezvous
	}

	if promListenPort := os.Getenv("PROM_LISTEN_PORT"); promListenPort != "" {
		Bridge.PromListenPort = &promListenPort
	}

	return nil
}

func (c *BridgeChain) GetEthClient(skipUrl string) (client *ethclient.Client, url string, err error) {
	for _, url := range c.RpcUrls {
		if skipUrl != "" && len(c.RpcUrls) > 1 && url == skipUrl {
			continue
		}
		if client, err = dial(url); err != nil {
			logrus.WithFields(logrus.Fields{
				field.CainId: c.Id,
				field.EthUrl: url,
			}).Error(fmt.Errorf("can not connect to chain rpc on error: %w", err))
			continue
		} else {
			balance, err := client.BalanceAt(context.Background(), c.EcdsaAddress, nil)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					field.CainId:       c.Id,
					field.EcdsaAddress: c.EcdsaAddress.String(),
				}).Error(fmt.Errorf("get address balance error: %w", err))
			}
			if balance.Cmp(big.NewInt(0)) <= 0 {
				logrus.WithFields(logrus.Fields{
					field.CainId:       c.Id,
					field.EcdsaAddress: c.EcdsaAddress.String(),
				}).Error(fmt.Errorf("node balance on chain [%d] is low", c.Id))
				return nil, url, fmt.Errorf("you balance on your chain [%d] wallet [%s]: %s to start node",
					c.Id, c.EcdsaAddress.String(), balance.String())

			}

			return client, url, nil
		}
	}
	return nil, "", fmt.Errorf("connection to all rpc url for chain [%d]  failed", c.Id)
}
