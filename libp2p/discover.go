package libp2p

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string, period time.Duration) {
	var routingDiscovery = discovery.NewRoutingDiscovery(dht)

	discovery.Advertise(ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := discovery.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				logrus.Fatal(err)
			}

			for _, p := range peers {
				if p.ID == h.ID() {
					continue
				}
				if h.Network().Connectedness(p.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, p.ID)
					fmt.Printf("Connected to peer %s\n", p.ID.Pretty())
					if err != nil {
						continue
					}
				}
			}
		}
	}
}
