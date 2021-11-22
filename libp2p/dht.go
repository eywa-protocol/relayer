package libp2p

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

func NewDHT(ctx context.Context, host host.Host, bootstrapAddrs []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {

	if len(bootstrapAddrs) == 0 {
		return nil, errors.New("empty bootstrap peers")
	}

	bootstrapPeers := make([]peer.AddrInfo, 0, len(bootstrapAddrs))
	for _, peerAddr := range bootstrapAddrs {
		if peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr); err != nil {
			err := fmt.Errorf("bootstrap peer AddrInfoFromP2pAddr error: %w", err)
			logrus.Error(err)
			return nil, err
		} else {
			bootstrapPeers = append(bootstrapPeers, *peerInfo)
		}
	}

	options := []dht.Option{
		dht.Mode(dht.ModeClient),
		dht.Mode(dht.ModeClient),
		dht.BootstrapPeers(bootstrapPeers...),
	}

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {

		return nil, fmt.Errorf("can not bootstrap DHT on error: %w", err)
	}

	mx := new(sync.Mutex)
	connected := false
	retryPeers := make([]peer.AddrInfo, 0, len(bootstrapAddrs))
	var wg sync.WaitGroup
	for _, bootstrapPeer := range bootstrapPeers {
		logrus.Printf("Bootstrap peer from DHT table: %s", bootstrapPeer)
		if host.ID().Pretty() != bootstrapPeer.ID.Pretty() {
			wg.Add(1)
			go func(peerInfo peer.AddrInfo) {
				defer wg.Done()
				if err := host.Connect(ctx, peerInfo); err != nil {
					logrus.WithFields(peerInfo.Loggable()).Error(fmt.Errorf("connecting to node error: %w", err))
					mx.Lock()
					retryPeers = append(retryPeers, peerInfo)
					mx.Unlock()
				} else {
					logrus.Infof("Connection established with bootstrap node: %q", peerInfo)
					mx.Lock()
					connected = true
					mx.Unlock()
				}
			}(bootstrapPeer)
		}
	}
	wg.Wait()

	// todo: move to configuration
	retryCount := 10
	retryTimeout := 120 * time.Second

	if !connected {
		for i := 0; i < retryCount; i++ {
			time.Sleep(retryTimeout)
			logrus.Infof("reconnect try %d of %d", i, retryCount)
			for _, retryPeer := range retryPeers {
				wg.Add(1)
				go func(peerInfo peer.AddrInfo) {
					defer wg.Done()
					if err := host.Connect(ctx, peerInfo); err != nil {
						logrus.Errorf("Error while connecting to node %q: %-v", peerInfo, err)
					} else {
						logrus.Infof("Connection established with bootstrap node: %q", peerInfo)
						mx.Lock()
						connected = true
						mx.Unlock()
					}
				}(retryPeer)
			}
			wg.Wait()
			if connected {
				break
			}
		}
	}

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
