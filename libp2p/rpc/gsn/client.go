package gsn

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/utils"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type ForwarderClientNode interface {
	GetDht() *dht.IpfsDHT
	GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error)
	GetForwarderAddress(chainId *big.Int) (common.Address, error)
}

type Client struct {
	ctx               context.Context
	host              host.Host
	node              ForwarderClientNode
	discoveryInterval time.Duration
	routingDiscovery  *discovery.RoutingDiscovery
	cid               cid.Cid
	rpcClient         *rpc.Client
	mx                *sync.Mutex
	gsnPeerId         peer.ID
}

func NewClient(ctx context.Context, host host.Host, node ForwarderClientNode, discoveryInterval time.Duration) (*Client, error) {
	if protocolCid, err := utils.ProtocolToCid(ProtocolId); err != nil {

		return nil, err
	} else {
		c := &Client{
			ctx:               ctx,
			host:              host,
			node:              node,
			discoveryInterval: discoveryInterval,
			routingDiscovery:  discovery.NewRoutingDiscovery(node.GetDht()),
			cid:               protocolCid,
			rpcClient:         rpc.NewClient(host, ProtocolId),
			mx:                new(sync.Mutex),
		}
		go c.discovery()

		return c, nil
	}
}

func (c *Client) getGsnPeerId() (peer.ID, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if c.gsnPeerId == "" {
		return "", errors.New("gsn node not discovered")
	} else {
		return c.gsnPeerId, nil
	}
}

func (c *Client) WaitForDiscoveryGsn(timeout time.Duration) error {
	var err error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	startTime := time.Now()
	stopTime := startTime.Add(timeout)
	go func() {
		defer wg.Done()
		for {
			if _, err = c.getGsnPeerId(); err == nil {
				logrus.Infof("gsn discover time: %s", time.Since(startTime).String())
				return
			} else if time.Now().After(stopTime) {
				err = errors.New("wait for discover gsn peer timed out")
				return
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
	wg.Wait()
	return err
}

func (c *Client) discovery() {
	ticker := time.NewTicker(c.discoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if c.gsnPeerId != "" && c.host.Network().Connectedness(c.gsnPeerId) == network.Connected {
				continue
			}
			gsnCh := c.routingDiscovery.FindProvidersAsync(c.ctx, c.cid, 1)
			for gsnPeer := range gsnCh {
				if c.host.Network().Connectedness(gsnPeer.ID) != network.Connected {
					_, err := c.host.Network().DialPeer(c.ctx, gsnPeer.ID)
					if err != nil {
						logrus.WithField(field.PeerId, gsnPeer.ID.Pretty()).
							Error(errors.New("connect to gsn peer was unsuccessful on discovery"))
					} else {
						logrus.Tracef("Discovery: connected to gsn peer %s", gsnPeer.ID.Pretty())
					}
				}
				if c.host.Network().Connectedness(gsnPeer.ID) == network.Connected {
					c.mx.Lock()
					if c.gsnPeerId == "" {
						logrus.Infof("gsn peer discovered: %s", gsnPeer.ID.Pretty())
					} else if c.gsnPeerId != gsnPeer.ID {
						logrus.Infof("gsn peer changed to: %s", gsnPeer.ID.Pretty())
					}
					c.gsnPeerId = gsnPeer.ID
					c.mx.Unlock()
				}
			}
		case <-c.ctx.Done():
			return
		}
	}

}

func (c *Client) Execute(chainId *big.Int, req wrappers.IForwarderForwardRequest, domainSeparator [32]byte, requestTypeHash [32]byte, suffixData []byte, sig []byte) (common.Hash, error) {

	var res ExecuteResult

	callReq := ExecuteRequest{
		ChainId:         chainId.String(),
		ForwardRequest:  NewRpcForwarderForwardRequest(&req),
		DomainSeparator: domainSeparator,
		RequestTypeHash: requestTypeHash,
		SuffixData:      suffixData,
		Signature:       sig,
	}

	logrus.Infof("forwarder request chainId: %s", chainId.String())

	if peerId, err := c.getGsnPeerId(); err != nil {

		return common.Hash{}, fmt.Errorf("get gsn peer ID error: %w", err)
	} else if err := c.rpcClient.CallContext(context.Background(), peerId, RpcService, RpcServiceFuncExecute, callReq, &res); err != nil {

		return common.Hash{}, fmt.Errorf("call rpc service [%s] method [%s] error: %w",
			RpcService, RpcServiceFuncExecute, err)
	} else {

		return res.TxHash, nil
	}
}

func (c *Client) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {

	return c.node.GetForwarder(chainId)
}

func (c *Client) GetForwarderAddress(chainId *big.Int) (common.Address, error) {

	return c.node.GetForwarderAddress(chainId)
}
