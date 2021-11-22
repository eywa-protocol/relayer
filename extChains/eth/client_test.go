package eth

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type clientList []Client

func (c clientList) Close() {
	for _, client := range c {
		client.Close()
	}
}

// Before run test you can :
// run from scripts ./1_clean_nodes.sh && ./2_redeploy_contracts.sh local && ./3_build_bsn.sh local && .4_build_config.sh
// stop ganache_net1 container
// after run test and see in console "wait for start ganache for 1111" start ganache_net1 container
func TestClient(t *testing.T) {
	err := config.LoadBridgeConfig("../../.data/bridge.yaml", false)
	if err != nil {
		logrus.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clients := make(clientList, 0, len(config.Bridge.Chains))
	defer clients.Close()

	txs := make([]string, 0, len(config.Bridge.Chains)*3)
	mx := new(sync.Mutex)
	eventHandler := func(event *wrappers.BridgeOracleRequest, srcChainId *big.Int) {
		mx.Lock()
		fmt.Printf("bridge oracle request received from %s to %s  tx: %s\n",
			srcChainId.String(), event.Chainid, event.Raw.TxHash.Hex())
		txs = append(txs, event.Raw.TxHash.Hex())
		mx.Unlock()
	}
	for _, chain := range config.Bridge.Chains {
		cfg := &Config{
			CallTimeout: 0,
			DialTimeout: 0,
			Id:          chain.Id,
			Urls:        chain.RpcUrls,
		}
		if client, err := NewClient(ctx, cfg); err != nil {

			fmt.Printf("create client error: %v\n", err)
			break
		} else if contractWatcher, err := NewOracleRequestWatcher(chain.BridgeAddress, eventHandler); err != nil {

			fmt.Printf("create watcher error: %v\n", err)
			break
		} else {
			client.AddWatcher(contractWatcher)
			clients = append(clients, client)
		}
	}
	if len(clients) != len(config.Bridge.Chains) {
		fmt.Println("some clients not initialized")
		t.Failed()
		return
	}

	fmt.Printf("wait for start ganache for %v\n", config.Bridge.Chains[0].Id)
	for {
		if chainId, err := clients[0].ChainID(ctx); err == nil && chainId != nil {
			fmt.Printf("end wait for start ganache for %v\n", config.Bridge.Chains[0].Id)
			break
		} else {
			time.Sleep(3 * time.Second)
		}
	}

	sendTxs := make([]string, 0, len(config.Bridge.Chains)*3)
	for i := 0; i < 3; i++ {

		for ixFrom, fromClient := range clients {
			var toClient Client
			toIx := 0
			if ixFrom < (len(clients) - 1) {
				toIx = ixFrom + 1
			}
			toClient = clients[toIx]

			testData := new(big.Int).SetUint64(rand.Uint64())

			dexPoolFrom := config.Bridge.Chains[ixFrom].DexPoolAddress
			dexPoolTo := config.Bridge.Chains[toIx].DexPoolAddress
			bridgeTo := config.Bridge.Chains[toIx].BridgeAddress
			pKeyFrom := config.Bridge.Chains[ixFrom].EcdsaKey

			txOptsFrom, err := fromClient.CallOpt(pKeyFrom)
			if err != nil {
				fmt.Printf("can not get call opt on error: %v", err)
				t.Failed()
				return
			}
			dexPoolFromContract, err := wrappers.NewMockDexPool(dexPoolFrom, fromClient)

			chainIdTo, err := toClient.ChainID(ctx)
			if err != nil {
				fmt.Printf("can not get chain id to on error: %v", err)
				t.Failed()
				return
			}
			tx, err := dexPoolFromContract.SendRequestTestV2(txOptsFrom,
				testData,
				dexPoolTo,
				bridgeTo,
				chainIdTo,
			)
			if err != nil {
				fmt.Printf("can not send tx on error: %v\n", err)
				t.Failed()
				return
			}
			fmt.Printf("tx send: %s\n", tx.Hash().Hex())
			_, _ = fromClient.WaitTransaction(tx.Hash())
			sendTxs = append(sendTxs, tx.Hash().Hex())
		}
	}
	mx.Lock()
	assert.Equal(t, len(sendTxs), len(txs))
	mx.Unlock()

}
