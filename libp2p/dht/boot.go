package dht

import (
	"context"
	"flag"
	"fmt"
	"github.com/DigiU-Lab/p2p-bridge/keys"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

func NewDHTBootPeer(keyFile string, port int) (err error) {
	logrus.Printf("keyFile %s port %d", keyFile, port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//настроить host который option
	listenAddr := libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	identify, err := identityFromKey(keyFile)
	if err != nil {
		return
	}
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		dualDHT, err := ddht.New(ctx, h, ddht.DHTOption(dht.Mode(dht.ModeServer), dht.ProtocolPrefix("/myapp"))) //в качестве dhtServer
		_ = dualDHT.Bootstrap(ctx)

		go func() {
			ticker := time.NewTicker(time.Second * 3)
			time.Sleep(time.Second)
			for {
				logrus.Printf("***** Вывод синхронизации RoutingTable ******\n")
				dualDHT.LAN.RoutingTable().Print()
				peerIds := dualDHT.LAN.RoutingTable().GetPeerInfos()
				for _, peerId := range peerIds {
					logrus.Printf("peerInfo AddedAt %v\n PeerID: %v\n LastSuccessfulOutboundQueryAt: %v\n LastUsefulAt: %v\n", peerId.AddedAt, peerId.Id, peerId.LastSuccessfulOutboundQueryAt, peerId.LastUsefulAt)
				}
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

	WriteHostAddrToConfig(host, "NewDHTBootPeer.env")
	sk := host.Peerstore().PrivKey(host.ID())
	bz, _ := crypto.MarshalPrivateKey(sk)
	logrus.Printf("KEY %v", bz)

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

func WriteHostAddrToConfig(host2 host.Host, filename string) {
	for i, addr := range host2.Addrs() {
		if i == 1 {
			nodeURL := fmt.Sprintf("%s/p2p/%s", addr, host2.ID().Pretty())
			logrus.Printf("Node Address: %s\n", nodeURL)
			err := ioutil.WriteFile(filename, []byte(nodeURL), 0777)
			if err != nil {
				logrus.Fatal(err)
			}
		}

	}
}

func NewPeer(keyFile string, port int, bootstrapPeers []multiaddr.Multiaddr) (err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var DigiUDefaultBootstrapPeers []multiaddr.Multiaddr

	for _, s := range []string{
		"/ip4/172.20.128.55/tcp/6666/p2p/QmYszwAKLoe1WjWutrc1tCCCKyHKBpr1a26LuESLqgcjHi",
		"/ip4/127.0.0.1/tcp/6666/p2p/QmdDKbtGuRwh5LmPGjRhjhMb7ivEi9autzpCeazYQnzZnw",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			return err
		}
		DigiUDefaultBootstrapPeers = append(DigiUDefaultBootstrapPeers, ma)
	}

	if len(bootstrapPeers) == 0 {
		bootstrapPeers = DigiUDefaultBootstrapPeers
	}

	// устанавка routingDiscovery
	var dualDHT *ddht.DHT
	var routingDiscovery *discovery.RoutingDiscovery
	routing := libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		var err error
		dualDHT, err = ddht.New(ctx, host, ddht.DHTOption(dht.ProtocolPrefix("/myapp"), dht.Mode(dht.ModeServer)))
		//,dht.ProtocolPrefix("/myapp")
		routingDiscovery = discovery.NewRoutingDiscovery(dualDHT)
		//_ = dualDHT.Bootstrap(ctx)
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			time.Sleep(time.Second)
			for {
				logrus.Printf("*****Выход синхронизации RoutingTable******\n")
				dualDHT.LAN.RoutingTable().Print()
				<-ticker.C
			}
		}()
		return dualDHT, err
	})

	identify, err := identityFromKey(keyFile)
	if err != nil {
		return
	}

	listenAddr := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")

	host, err := libp2p.New(
		ctx,
		routing,
		identify,
		listenAddr,
		//libp2p.NATPortMap(),
	)
	if err != nil {
		return
	}
	host.SetStreamHandler(protocol.ID("/chat/1.0"), handleStream)
	for _, addr := range host.Addrs() {
		logrus.Printf("Address: %s\n", addr)
	}
	logrus.Printf("local id is: %s\n", host.ID().Pretty())

	if err := dualDHT.Bootstrap(ctx); err != nil {
		return err
	}
	logrus.Println("ПЩШТП ЕЩ CONNECTING TO PEERS  --------------- bootstrapPeers")
	var wg sync.WaitGroup
	for _, maddr := range bootstrapPeers {
		logrus.Println("CONNECTING TO PEERS  --------------- bootstrapPeers")
		wg.Add(1)
		peerInfo, _ := peer.AddrInfoFromP2pAddr(maddr)
		logrus.WithField(" --------------- peerInfo", peerInfo)
		go func() {
			defer wg.Done()
			logrus.Printf("Пробую подключиться к :%s\n", peerInfo.ID.Pretty())
			if err := host.Connect(ctx, *peerInfo); err != nil {
				logrus.Printf("Сбой соединения :%s\n", peerInfo.ID.Pretty())
			} else {
				logrus.Printf("Успешно подключен к :%s\n", peerInfo.ID.Pretty())
			}
		}()
	}
	wg.Wait()
	logrus.Printf("Таблица маршрутизации\n")
	logrus.Printf("LAN:\n")
	dualDHT.LAN.RoutingTable().Print()
	logrus.Printf("WAN:\n")
	dualDHT.WAN.RoutingTable().Print()
	room := "room"
	//dualDHT.WAN.RoutingTable().NearestPeer()
	//Обработка чатов

	discovery.Advertise(ctx, routingDiscovery, "room")
	logrus.Printf("Успех будет room:%s транслироваться\n", room)
	select {}

	for {
		logrus.Printf("Начинай искать peers\n")
		peerInfos, err := discovery.FindPeers(ctx, routingDiscovery, "room")
		if len(peerInfos) != 0 {
			discovery.Advertise(ctx, routingDiscovery, room)
		} else {
			continue
		}

		logrus.Printf("peers:\n")
		for i, pe := range peerInfos {
			logrus.Printf("(%d):%s\n", i, pe.ID.Pretty())
		}
		if err != nil {
			panic(err)
		}
		for _, peerInfo := range peerInfos {
			logrus.Printf("найти peer:%s\n", peerInfo.ID.Pretty())
			if peerInfo.ID == host.ID() {
				continue
			}
			logrus.Printf("Пробую подключиться peer:%s\n", peerInfo.ID.Pretty())
			if err := host.Connect(ctx, peerInfo); err != nil {
				logrus.Printf(" соединение не удалось :%s\n", peerInfo.ID.Pretty())
				continue
			}
			logrus.Printf("Успешное соединение peer:%s\n", peerInfo.ID.Pretty())
			logrus.Printf("Попробуйте создать stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())
			stream, err := host.NewStream(ctx, peerInfo.ID, "/chat/1.0")
			if err != nil {
				panic(err)
			} else {
				logrus.Printf("Успешное создание stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())

			}
			go handleStream(stream)
		}
		time.Sleep(time.Minute * 2)
	}
}

func handleStream(stream network.Stream) {
	go logrus.Printf("Получите новый stream\n")
}

// A new type we need for writing a custom flag parser
type addrList []multiaddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (multiaddrs []multiaddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := multiaddr.NewMultiaddr(addrString)
		if err != nil {
			return multiaddrs, err
		}
		multiaddrs = append(multiaddrs, addr)
	}
	return
}

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
	var routingDiscovery = discovery.NewRoutingDiscovery(dht)

	discovery.Advertise(ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := discovery.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				logrus.Errorf("%v", err)
			}

			for _, p := range peers {
				if p.ID == h.ID() {
					continue
				}
				if h.Network().Connectedness(p.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, p.ID)
					logrus.Printf("Connected to peer %s\n", p.ID.Pretty())
					if err != nil {
						continue
					}
				}
			}
		}
	}
}

//func  Print(rt *kbucket.RoutingTable) {
//	fmt.Printf("Routing Table, bs = %d, Max latency = %d\n", rt.bucketsize, rt.maxLatency)
//	rt.tabLock.RLock()
//     rt.GetPeerInfos()
//	for i, b := range rt.buckets {
//		fmt.Printf("\tbucket: %d\n", i)
//
//		for e := b.list.Front(); e != nil; e = e.Next() {
//			p := e.Value.(*PeerInfo).Id
//			fmt.Printf("\t\t- %s %s\n", p.Pretty(), rt.metrics.LatencyEWMA(p).String())
//		}
//	}
//	rt.tabLock.RUnlock()
//}

/*func ConnectToPeers(peerIDs []peer.ID) {
	for _, id := range peerIDs {
		peer := n.Host.Peerstore().PeerInfo(id)
		err := n.Connect(n.ctx, peer)
		if err == nil {
			logrus.Println(n.Host.ID(), " Connected to ", peer.Addrs)
			connected++
		}
	}

	if connected > 0 {
		logrus.Println("Advertising")
		n.Router.Advertise(n.ctx, n.cfg.RendezvousString)
		close(n.bootstrapDone)
		break
	}
}*/

func identityFromKey(keyFile string) (identity libp2p.Option, err error) {
	privKey, err := keys.ReadHostKey(keyFile)
	if err != nil {
		logrus.Errorf("ERROR GETTING CERT %v", err)
		return
	}
	identity = libp2p.Identity(privKey)
	return
}
