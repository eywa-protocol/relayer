package config

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gopkg.in/yaml.v3"
	_ "gopkg.in/yaml.v3"
)

// Bridge is global object that holds all bridge configuration variables.
var Bridge Configuration

type Configuration struct {
	TickerInterval       time.Duration `yaml:"ticker_interval"`
	GsnDiscoveryInterval time.Duration `yaml:"gsn_discovery_interval"`
	GsnWaitDuration      time.Duration `yaml:"gsn_wait_duration"`
	UptimeReportInterval time.Duration
	UseGsn               bool
	Rendezvous           string         `yaml:"rendezvous"`
	RegChainId           uint64         `yaml:"reg_chain_id"`
	Chains               []*BridgeChain `yaml:"chains"`
	BootstrapAddrs       []string       `yaml:"bootstrap-addrs"`
	PromListenPort       *string        `yaml:"prom_listen_port"`
}

type BridgeChain struct {
	CallTimeout         time.Duration `yaml:"call_timeout"`
	DialTimeout         time.Duration `yaml:"dial_timeout"`
	BlockTimeout        time.Duration `yaml:"block_timeout"`
	Id                  uint64        `yaml:"id"`
	ChainId             *big.Int
	RpcUrls             []string `yaml:"rpc_urls"`
	EcdsaKeyString      string   `yaml:"ecdsa_key"`
	EcdsaKey            *ecdsa.PrivateKey
	EcdsaAddress        common.Address
	BridgeAddress       common.Address `yaml:"bridge_address"`
	NodeRegistryAddress common.Address `yaml:"node_registry_address"`
	DexPoolAddress      common.Address `yaml:"dex_pool_address"`
	ForwarderAddress    common.Address `yaml:"forwarder_address"`
	UseGsn              bool
}

func LoadBridgeConfig(path string, useGsn bool) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file [%s] error: %w",
			path, err)
	}
	if err := yaml.Unmarshal(data, &Bridge); err != nil {
		return fmt.Errorf("unmarshal config error: %w", err)
	}

	if Bridge.GsnWaitDuration <= 0 {

		Bridge.GsnWaitDuration = 60 * time.Second
	}

	if Bridge.GsnDiscoveryInterval <= 0 {

		Bridge.GsnDiscoveryInterval = 10 * time.Second
	}

	if Bridge.RegChainId <= 0 {
		Bridge.RegChainId = Bridge.Chains[0].Id
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

		if useGsn && !common2.AddressIsZero(chain.ForwarderAddress) {
			Bridge.UseGsn = true
			chain.UseGsn = true
		}
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

// func (c *BridgeChain) GetEthClient(skipUrl string, signerAddress common.Address) (client *ethclient.Client, url string, err error) {
// 	for _, url := range c.RpcUrls {
// 		if skipUrl != "" && len(c.RpcUrls) > 1 && url == skipUrl {
// 			continue
// 		}
// 		if client, err = ethclient.Dial(url); err != nil {
// 			logrus.WithFields(logrus.Fields{
// 				field.CainId: c.Id,
// 				field.EthUrl: url,
// 			}).Error(fmt.Errorf("can not connect to chain rpc on error: %w", err))
// 			continue
// 		} else {
// 			balance, err := client.BalanceAt(context.Background(), signerAddress, nil)
// 			if err != nil {
// 				logrus.WithFields(logrus.Fields{
// 					field.CainId:       c.Id,
// 					field.EcdsaAddress: signerAddress,
// 				}).Error(fmt.Errorf("get address balance error: %w", err))
// 			}
// 			if balance == big.NewInt(0) {
//
// 				return nil, url, fmt.Errorf("you balance on your chain [%d] wallet [%s]: %s to start node",
// 					c.Id, signerAddress.String(), balance.String())
//
// 			}
//
// 			return client, url, nil
// 		}
// 	}
// 	return nil, "", fmt.Errorf("connection to all rpc url for chain [%d]  failed", c.Id)
// }
