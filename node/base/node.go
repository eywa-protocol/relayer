package base

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
)

type Node struct {
	Ctx  context.Context
	Host host.Host
	Dht  *dht.IpfsDHT
}

func (n Node) InitDHT(bootstrapPeerAddrs []string) (dht *dht.IpfsDHT, err error) {

	bootstrapPeers := make([]multiaddr.Multiaddr, 0, len(bootstrapPeerAddrs))

	for _, addr := range bootstrapPeerAddrs {
		logrus.Infof("add bootstrap peer: %s", addr)
		nAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		bootstrapPeers = append(bootstrapPeers, nAddr)
	}
	logrus.Infof("bootstrap peers count: %d", len(bootstrapPeers))
	dht, err = libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers[:])
	if err != nil {
		return
	}
	return
}

func (n Node) GetDht() *dht.IpfsDHT {
	return n.Dht
}
