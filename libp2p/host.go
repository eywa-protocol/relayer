package libp2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multiaddr"
	"io"
	mrand "math/rand"
)

func NewHost(ctx context.Context, seed int64, keyFile string, port int) (host host.Host, err error) {
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}
	addr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if keyFile == "" {
		priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		if err != nil {
			return nil, err
		}

		host, err = libp2p.New(ctx,
			libp2p.ListenAddrs(addr),
			libp2p.Identity(priv),
		)

	} else {
		id, err := IdentityFromKey(keyFile)
		if err != nil {
			return nil, err
		}
		host, err = libp2p.New(ctx,
			libp2p.ListenAddrs(addr),
			id,
		)
	}
	return

}

func NewHostFromKeyFila(ctx context.Context, keyFile string, port int, address string) (host2 host.Host, err error) {
	if address == "" {
		address = "0.0.0.0"
	}
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", address, port))
	if err != nil {
		return
	}

	id, err := IdentityFromKey(keyFile)
	if err != nil {
		return
	}

	host2, err = libp2p.New(ctx,
		libp2p.ListenAddrs(addr),
		id)
	return
}
