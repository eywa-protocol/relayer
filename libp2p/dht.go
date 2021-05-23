package libp2p

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
)

func NewDHT(ctx context.Context, host host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	var options []dht.Option

	if len(bootstrapPeers) == 0 {
		options = append(options, dht.Mode(dht.ModeServer))
	}

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		if host.ID().Pretty() != peerinfo.ID.Pretty() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n := 0
				for {
					if n == 10 {
						break
					}
					n++

					if err := host.Connect(ctx, *peerinfo); err != nil {
						logrus.Errorf("Error while connecting to node %q: %-v", peerinfo, err)
					} else {
						logrus.Infof("Connection established with node: %q", peerinfo)
						break
					}
					time.Sleep(10 * time.Second)
					fmt.Println("Trying to connect after unsuccessful connect...")
				}
			}()
		}
	}
	wg.Wait()

	return kdht, nil
}

func WriteHostAddrToConfig(host2 host.Host, filename string) (nodeURL string) {
	for i, addr := range host2.Addrs() {
		if i == 0 {
			nodeURL = fmt.Sprintf("%s/p2p/%s", addr, host2.ID().Pretty())
			logrus.Infof("WriteHostAddrToConfig Node Address: %s", nodeURL)
			err := ioutil.WriteFile(filename, []byte(nodeURL), 0644)
			if err != nil {
				logrus.Fatal(err)
			}
		}
	}
	return
}

func GetAddrsFromHost(host2 host.Host) (nodeAddrs []string) {
	for _, addr := range host2.Addrs() {
		nodeURL := fmt.Sprintf("%s/p2p/%s", addr, host2.ID().Pretty())
		logrus.Tracef("Node Address: %s\n", nodeURL)
		nodeAddrs = append(nodeAddrs, nodeURL)
	}
	return
}
