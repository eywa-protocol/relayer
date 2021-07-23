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

// App is global object that holds all application level variables.
var App Configuration

type Configuration struct {
	TickerInterval       time.Duration `yaml:"ticker_interval"`
	UptimeReportInterval time.Duration
	Rendezvous           string   `yaml:"rendezvous"`
	Chains               []*Chain `yaml:"chains"`
	BootstrapAddrs       []string `yaml:"bootstrap-addrs"`
}

type Chain struct {
	Id              uint `yaml:"id"`
	ChainId         *big.Int
	EcdsaKeyString  string `yaml:"ecdsa_key"`
	EcdsaKey        *ecdsa.PrivateKey
	EcdsaAddress    common.Address
	BridgeAddress   common.Address `yaml:"bridge_address"`
	NodeListAddress common.Address `yaml:"node_list_address"`
	DexPoolAddress  common.Address `yaml:"dex_pool_address"`
	RpcUrls         []string       `yaml:"rpc_urls"`
}

func Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file [%s] error: %w",
			path, err)
	}
	if err := yaml.Unmarshal(data, &App); err != nil {
		return fmt.Errorf("unmarshal config error: %w", err)
	}

	// todo: change to 24 hours
	App.UptimeReportInterval = 5 * time.Minute

	for _, chain := range App.Chains {
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
		App.Rendezvous = rendezvous
	}

	return nil
}

func (c *Chain) GetEthClient() (client *ethclient.Client, err error) {
	for _, url := range c.RpcUrls {
		if client, err = ethclient.Dial(url); err != nil {
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
			if balance == big.NewInt(0) {

				return nil, fmt.Errorf("you balance on your chain [%d] wallet [%s]: %s to start node",
					c.Id, c.EcdsaAddress.String(), balance.String())

			}

			return client, nil
		}
	}
	return nil, fmt.Errorf("connection to all rpc url for chain [%d]  failed", c.Id)
}
