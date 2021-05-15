package main

//
//func main() {
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	var bootstrapPeers addrList
//	var DigiUDefaultBootstrapPeers []multiaddr.Multiaddr
//
//	for _, s := range []string{
//		"/ip4/172.20.128.55/tcp/6666/p2p/QmYszwAKLoe1WjWutrc1tCCCKyHKBpr1a26LuESLqgcjHi",
//		"/ip4/127.0.0.1/tcp/6666/p2p/QmdDKbtGuRwh5LmPGjRhjhMb7ivEi9autzpCeazYQnzZnw",
//	} {
//		ma, err := multiaddr.NewMultiaddr(s)
//		if err != nil {
//			panic(err)
//		}
//		DigiUDefaultBootstrapPeers = append(DigiUDefaultBootstrapPeers, ma)
//	}
//
//	//firstMultiAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/6666/p2p/QmdDKbtGuRwh5LmPGjRhjhMb7ivEi9autzpCeazYQnzZnw")
//
//	//DigiUDefaultBootstrapPeers = append(DigiUDefaultBootstrapPeers, firstMultiAddr)
//	flag.Var(&bootstrapPeers, "b", "")
//	flag.Parse()
//	if len(bootstrapPeers) == 0 {
//		bootstrapPeers = DigiUDefaultBootstrapPeers
//	}
//
//	// устанавка routingDiscovery
//	var dualDHT *ddht.DHT
//	var routingDiscovery *discovery.RoutingDiscovery
//	routing := libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
//		var err error
//		dualDHT, err = ddht.New(ctx, host, ddht.DHTOption(dht.ProtocolPrefix("/myapp"), dht.Mode(dht.ModeServer)))
//		//,dht.ProtocolPrefix("/myapp")
//		routingDiscovery = discovery.NewRoutingDiscovery(dualDHT)
//		//_ = dualDHT.Bootstrap(ctx)
//		go func() {
//			ticker := time.NewTicker(time.Second * 15)
//			time.Sleep(time.Second)
//			for {
//				logrus.Printf("*****Выход синхронизации RoutingTable******\n")
//				dualDHT.LAN.RoutingTable().Print()
//				<-ticker.C
//			}
//		}()
//		return dualDHT, err
//	})
//
//	privKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
//	identify := libp2p.Identity(privKey)
//
//	listenAddr := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0")
//
//	//创建host
//	host, err := libp2p.New(
//		ctx,
//		routing,
//		identify,
//		listenAddr,
//		//libp2p.NATPortMap(),
//	)
//	if err != nil {
//		panic(err)
//	}
//	host.SetStreamHandler(protocol.ID("/chat/1.0"), handleStream)
//	for _, addr := range host.Addrs() {
//		logrus.Printf("Address: %s\n", addr)
//	}
//	logrus.Printf("local id is: %s\n", host.ID().Pretty())
//
//	//连接到bootstrap peer
//	if err := dualDHT.Bootstrap(ctx); err != nil {
//		panic(err)
//	}
//
//	var wg sync.WaitGroup
//	for _, maddr := range bootstrapPeers {
//		wg.Add(1)
//		peerInfo, _ := peer.AddrInfoFromP2pAddr(maddr)
//		go func() {
//			defer wg.Done()
//			logrus.Printf("Пробую подключиться к :%s\n", peerInfo.ID.Pretty())
//			if err := host.Connect(ctx, *peerInfo); err != nil {
//				logrus.Printf("Сбой соединения :%s\n", peerInfo.ID.Pretty())
//			} else {
//				logrus.Printf("Успешно подключен к :%s\n", peerInfo.ID.Pretty())
//			}
//		}()
//	}
//	wg.Wait()
//
//	logrus.Printf("Таблица маршрутизации\n")
//	logrus.Printf("LAN:\n")
//	dualDHT.LAN.RoutingTable().Print()
//	logrus.Printf("WAN:\n")
//	dualDHT.WAN.RoutingTable().Print()
//	room := "room"
//
//	//Обработка чатов
//
//	discovery.Advertise(ctx, routingDiscovery, "room")
//	logrus.Printf("Успех будет room:%s транслироваться\n", room)
//	select {}
//
//	for {
//		logrus.Printf("Начинай искать peers\n")
//		peerInfos, err := discovery.FindPeers(ctx, routingDiscovery, "room")
//		if len(peerInfos) != 0 {
//			discovery.Advertise(ctx, routingDiscovery, room)
//		} else {
//			continue
//		}
//
//		logrus.Printf("peers:\n")
//		for i, pe := range peerInfos {
//			logrus.Printf("(%d):%s\n", i, pe.ID.Pretty())
//		}
//		if err != nil {
//			panic(err)
//		}
//		for _, peerInfo := range peerInfos {
//			logrus.Printf("найти peer:%s\n", peerInfo.ID.Pretty())
//			if peerInfo.ID == host.ID() {
//				continue
//			}
//			logrus.Printf("Попробуйте подключиться peer:%s\n", peerInfo.ID.Pretty())
//			if err := host.Connect(ctx, peerInfo); err != nil {
//				logrus.Printf("连接peer失败:%s\n", peerInfo.ID.Pretty())
//				continue
//			}
//			logrus.Printf("Успешное соединение peer:%s\n", peerInfo.ID.Pretty())
//			logrus.Printf("Попробуйте создать stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())
//			stream, err := host.NewStream(ctx, peerInfo.ID, "/chat/1.0")
//			if err != nil {
//				panic(err)
//			} else {
//				logrus.Printf("Успешное создание stream:%s<------>%s\n", host.ID().Pretty(), peerInfo.ID.Pretty())
//
//			}
//			go handleStream(stream)
//		}
//		time.Sleep(time.Minute * 2)
//	}
//}
//
//func handleStream(stream network.Stream) {
//	go logrus.Printf("Получите новый stream\n")
//}
//
//// A new type we need for writing a custom flag parser
//type addrList []multiaddr.Multiaddr
//
//func (al *addrList) String() string {
//	strs := make([]string, len(*al))
//	for i, addr := range *al {
//		strs[i] = addr.String()
//	}
//	return strings.Join(strs, ",")
//}
//
//func (al *addrList) Set(value string) error {
//	addr, err := multiaddr.NewMultiaddr(value)
//	if err != nil {
//		return err
//	}
//	*al = append(*al, addr)
//	return nil
//}
//
//func StringsToAddrs(addrStrings []string) (multiaddrs []multiaddr.Multiaddr, err error) {
//	for _, addrString := range addrStrings {
//		addr, err := multiaddr.NewMultiaddr(addrString)
//		if err != nil {
//			return multiaddrs, err
//		}
//		multiaddrs = append(multiaddrs, addr)
//	}
//	return
//}
//
//func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
//	var routingDiscovery = discovery.NewRoutingDiscovery(dht)
//
//	discovery.Advertise(ctx, routingDiscovery, rendezvous)
//
//	ticker := time.NewTicker(time.Second * 1)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-ticker.C:
//
//			peers, err := discovery.FindPeers(ctx, routingDiscovery, rendezvous)
//			if err != nil {
//				logrus.Errorf("%v", err)
//			}
//
//			for _, p := range peers {
//				if p.ID == h.ID() {
//					continue
//				}
//				if h.Network().Connectedness(p.ID) != network.Connected {
//					_, err = h.Network().DialPeer(ctx, p.ID)
//					logrus.Printf("Connected to peer %s\n", p.ID.Pretty())
//					if err != nil {
//						continue
//					}
//				}
//			}
//		}
//	}
//}
