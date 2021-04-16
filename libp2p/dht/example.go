package dht

import (
	"context"
	"github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	h2 "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

var BootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/104.236.176.52/tcp/4001/ipfs/QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/162.243.248.213/tcp/4001/ipfs/QmSoLueR4xBeUbY9WZ9xGUUxunbKWcrNFTDAadQJmocnWm",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
}

func config(cfg *libp2p.Config) error {
	addr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	if err != nil {
		return err
	}
	cfg.ListenAddrs = []multiaddr.Multiaddr{
		addr,
	}
	return libp2p.Defaults(cfg)
}

func peerFactory() (*dht.IpfsDHT, h2.Host) {

	//Create host
	h, err := libp2p.New(context.Background(), config)
	if err != nil {
		panic(err)
	}

	//Create DHT
	d := dht.NewDHTClient(context.Background(), h, datastore.NewMapDatastore())

	for _, p := range BootstrapPeers {
		//Bootstrap object
		ma, err := multiaddr.NewMultiaddr(p)
		if err != nil {
			panic(err)
		}

		info, err := peerstore.InfoFromP2pAddr(ma)
		err = h.Connect(context.Background(), *info)
		if err != nil {
			panic(err)
		}
	}

	//Exit on DHT bootstrap error
	if err := d.Bootstrap(context.Background()); err != nil {
		panic(err)
	}

	return d, h

}

func Newmain() {

	dhtAlice, _ := peerFactory()
	_, hostBob := peerFactory()

	logrus.Println(dhtAlice.FindPeer(context.Background(), hostBob.ID()))

}
