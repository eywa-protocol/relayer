package libp2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

func NewHostFromKeyFile(ctx context.Context, keyFile string, port int, address string) (host2 host.Host, err error) {
	if address == "" {
		address = "0.0.0.0"
	}
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", address, port))
	if err != nil {
		return
	}

	privKey, err := common.ReadHostKey(keyFile)
	if err != nil {
		logrus.Errorf("ERROR GETTING CERT %v", err)
		return
	}

	return NewHost(ctx, privKey, addr)
}

func NewHost(ctx context.Context, prvKey crypto.PrivKey, multiAddr multiaddr.Multiaddr) (host2 host.Host, err error) {

	return libp2p.New(
		ctx,
		libp2p.ListenAddrs(multiAddr),
		libp2p.Identity(prvKey),
		// libp2p.Security(tls.ID, tls.New),
		// libp2p.EnableNATService(),
		// libp2p.ForceReachabilityPublic(),
	)
}
