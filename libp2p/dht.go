package libp2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func NewDHT(ctx context.Context, host host.Host) (*dht.IpfsDHT, error) {
	var options []dht.Option

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	bootstrapPeers := dht.DefaultBootstrapPeers

	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		logrus.Printf("Bootstrap peer from DHT table: %s", peerinfo)
		if host.ID().Pretty() != peerinfo.ID.Pretty() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := host.Connect(ctx, *peerinfo); err != nil {
					logrus.Errorf("Error while connecting to node %q: %-v", peerinfo, err)
				} else {
					logrus.Infof("Connection established with node: %q", peerinfo)

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
