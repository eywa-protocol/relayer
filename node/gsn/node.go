package gsn

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type Node struct {
	base.Node
	cMx     *sync.Mutex
	Clients map[string]Client
}

func RunNode(name, keysPath, listen string, port uint) (err error) {

	nodeKey, err := common2.GetOrSaveECDSAKey(keysPath, name)
	if err != nil {
		panic(err)
	}

	logrus.Tracef("keyfile %v", keysPath+"/"+name+"-ecdsa.key")

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listen, port))
	if err != nil {
		panic(err)
	}

	h, err := libp2p.NewHostFromKey(nodeKey, multiAddr)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	n := &Node{
		Node:    base.Node{Ctx: ctx, Host: h},
		cMx:     new(sync.Mutex),
		Clients: make(map[string]Client, len(config.Gsn.Chains)),
	}

	logrus.Print(len(config.Gsn.Chains), " chains Length")
	for _, chain := range config.Gsn.Chains {
		logrus.Print("CHAIN ", chain, "chain")

		client, err := NewClient(chain, "")
		if err != nil {
			return fmt.Errorf("init chain[%d] node client error: %w", chain.Id, err)
		}

		if reflect.DeepEqual(client, ethclient.Client{}) {
			return fmt.Errorf("init chain [%d] client failed", chain.Id)
		}
		if _, ok := n.Clients[client.ChainCfg.ChainId.String()]; ok {
			return fmt.Errorf("init duplicate  chain[%d] node client chainId:[%s] error %w", chain.Id, client.ChainCfg.ChainId, err)
		}
		n.Clients[client.ChainCfg.ChainId.String()] = client
	}

	n.Dht, err = n.InitDHT(config.Gsn.BootstrapAddrs)
	if err != nil {
		return fmt.Errorf("init DHT error: %w", err)
	}

	if _, err := gsn.NewServer(n.Ctx, n.Host, n); err != nil {
		return fmt.Errorf("init gsn rpc service error: %w", err)
	}

	runa.Host(h, cancel, &sync.WaitGroup{})

	return
}

func (n *Node) GetNodeClientOrRecreate(chainId *big.Int) (Client, bool, error) {
	n.cMx.Lock()
	clientRecreated := false
	client, ok := n.Clients[chainId.String()]
	if !ok {
		n.cMx.Unlock()
		return Client{}, false, errors.New("eth client for chain ID not found")
	}
	n.cMx.Unlock()

	netChainId, err := client.EthClient.ChainID(n.Ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			field.CainId: chainId,
		}).Error(fmt.Errorf("recreate network client on error: %w", err))

		client, err = NewClient(client.ChainCfg, client.currentUrl)
		if err != nil {
			err = fmt.Errorf("can not create client on error:%w", err)
			return Client{}, false, err
		} else {
			// replace network client in clients map
			clientRecreated = true
			n.cMx.Lock()
			n.Clients[chainId.String()] = client
			n.cMx.Unlock()
		}
	}

	if netChainId != nil && netChainId.Cmp(chainId) != 0 {
		err = errors.New("chain id not match to network")
		logrus.WithFields(logrus.Fields{
			field.CainId:         chainId,
			field.NetworkChainId: netChainId,
		}).Error(err)
		return Client{}, false, err
	}

	return client, clientRecreated, nil
}

func (n Node) GetNodeClient(chainId *big.Int) (Client, error) {
	n.cMx.Lock()
	defer n.cMx.Unlock()

	client, ok := n.Clients[chainId.String()]
	if !ok {
		return Client{}, errors.New("eth client for chain ID not found")
	}

	return client, nil
}

func (n *Node) GetOwner(chainId *big.Int) (*bind.TransactOpts, error) {
	if client, err := n.GetNodeClient(chainId); err != nil {

		return nil, fmt.Errorf("get forwarder owner for chain[%s] error: %w", chainId.String(), err)
	} else {

		return client.owner, nil
	}
}

func (n *Node) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {
	if client, _, err := n.GetNodeClientOrRecreate(chainId); err != nil {

		return nil, fmt.Errorf("get forwarder owner for chain[%s] error: %w", chainId.String(), err)
	} else {

		return client.forwarder, nil
	}
}
