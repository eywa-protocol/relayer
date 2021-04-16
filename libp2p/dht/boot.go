package dht

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/sirupsen/logrus"
	"github.com/sivo4kin/ea-starter/keys"
	"os"
	"time"
)

func NewDHTBootPeer(keyFile string, port int) (err error) {
	logrus.Printf("keyFile %s port %d", keyFile, port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//настроить host который option
	listenAddr := libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	privKey, err := keys.ReadHostKey(keyFile)
	if err != nil {
		logrus.Errorf("ERROR GETTING CERT %v", err)
		//panic(err)
		return
	}
	identify := libp2p.Identity(privKey)
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		dualDHT, err := ddht.New(ctx, h, ddht.DHTOption(dht.Mode(dht.ModeServer), dht.ProtocolPrefix("/myapp"))) //в качестве dhtServer
		//,dht.ProtocolPrefix("/myapp")
		_ = dualDHT.Bootstrap(ctx)

		go func() {
			ticker := time.NewTicker(time.Second * 15)
			time.Sleep(time.Second)
			for {
				logrus.Printf("***** Вывод синхронизации RoutingTable ******\n")
				dualDHT.LAN.RoutingTable().Print()
				<-ticker.C
			}
		}()
		return dualDHT, err
	})

	host, err := libp2p.New(
		ctx,
		identify,
		routing,
		listenAddr,
		//libp2p.NATPortMap(),
	)
	if err != nil {
		//panic(err)
		return
	}
	for i, addr := range host.Addrs() {
		if i == 0 {
			nodeURL := fmt.Sprintf("%s/p2p/%s", addr, host.ID().Pretty())
			logrus.Printf("Node Address: %s\n", nodeURL)
			setBootNodeURLEnv(nodeURL)
		}

	}

	host.SetStreamHandler(dht.ProtocolDHT, func(stream network.Stream) {
		logrus.Printf("handling %sv\n", stream)
	})

	select {}
}

func setFlags(ctx context.Context, port *int) {
	flag.IntVar(port, "port", 6666, "")
}

func setBootNodeURLEnv(botNodeURL string) {
	err := os.Setenv("P2P_URL", botNodeURL)
	if err != nil {
		logrus.Errorf("err %v", err)
	}
}
