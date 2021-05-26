package node

import (
	"context"
	"errors"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

func loadNodeConfig(path string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Tracef("started in directory %s", dir)
	err = config.LoadConfigAndArgs(path)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	hostName, _ := os.Hostname()

	keysList := os.Getenv("ECDSA_KEY_1")
	keys := strings.Split(keysList, ",")
	keysList2 := os.Getenv("ECDSA_KEY_2")
	keys2 := strings.Split(keysList2, ",")

	strNum := strings.TrimPrefix(hostName, "p2p-bridge_node_")
	strNum = strings.TrimRightFunc(strNum, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	nodeHostId, _ := strconv.Atoi(strNum)

	c1, c2, err := getEthClients()
	if err != nil {
		logrus.Fatal(err)

	}

	var getRandomKeyForTestIfNoFunds func()
	getRandomKeyForTestIfNoFunds = func() {
		config.Config.ECDSA_KEY_1 = keys[nodeHostId]
		config.Config.ECDSA_KEY_2 = keys2[nodeHostId]
		balance1, err := c1.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_1), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		balance2, err := c2.BalanceAt(context.Background(), common2.AddressFromPrivKey(config.Config.ECDSA_KEY_2), nil)
		if err != nil {
			logrus.Fatal(err)
		}

		if balance1 == big.NewInt(0) || balance2 == big.NewInt(0) {
			logrus.Errorf("you need balance on your wallets 1: %d 2: %d to start node", balance1, balance2)
			//if nodeHostId == 0 || nodeHostId > len(keys)-1 {
			rand.Seed(time.Now().UnixNano())
			nodeHostId = rand.Intn(len(keys))
			//}
			getRandomKeyForTestIfNoFunds()
		}

	}

	config.Config.ECDSA_KEY_1 = keys[nodeHostId]
	config.Config.ECDSA_KEY_2 = keys2[nodeHostId]

	if config.Config.ECDSA_KEY_1 == "" || config.Config.ECDSA_KEY_2 == "" {
		panic(errors.New("you need key to start node"))
	}

	logrus.Tracef("hostName %s nodeHostId %v key1 %v key2 %v", hostName, nodeHostId, config.Config.ECDSA_KEY_1, config.Config.ECDSA_KEY_2)

	return
}

func NodeInit(path, name string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	if common2.FileExists("keys/" + name + "-ecdsa.key") {
		return errors.New("node allready registered! ")
	}

	err = common2.GenAndSaveECDSAKey(name)
	if err != nil {
		panic(err)
	}

	_, _, err = common2.GenAndSaveBN256Key(name)
	if err != nil {
		panic(err)
	}

	blsAddr, pub, err := common2.GenAndSaveBN256Key(name)
	if err != nil {
		return
	}

	err = common2.GenAndSaveECDSAKey(name)
	if err != nil {
		return
	}

	logrus.Tracef("keyfile %v", "keys/"+name+"-ecdsa.key")

	h, err := libp2p.NewHostFromKeyFila(context.Background(), "keys/"+name+"-ecdsa.key", 0, "")
	if err != nil {
		panic(err)
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, "keys/"+name+"-peer.env")
	c1, c2, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Infof("nodelist1 blsAddress: %v", blsAddr)
	pKey1, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	err = common2.RegisterNode(c1, pKey1, common.HexToAddress(config.Config.NODELIST_NETWORK1), common.HexToAddress(config.Config.ECDSA_KEY_1), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network1 %v", err)
	}

	pKey2, err := common2.ToECDSAFromHex(config.Config.ECDSA_KEY_2)
	if err != nil {
		return
	}
	err = common2.RegisterNode(c2, pKey2, common.HexToAddress(config.Config.NODELIST_NETWORK2), common.HexToAddress(config.Config.ECDSA_KEY_2), []byte(nodeURL), []byte(pub), blsAddr)
	if err != nil {
		logrus.Errorf("error registaring node in network2 %v", err)
	}

	return
}

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logrus.Infof("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func NewNode(path, name string) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx: ctx,
	}
	//n.PublicKeys = make([]kyber.Point,0)

	n.pKey, err = common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Tracef("n.Config.PORT_1 %d", config.Config.PORT_1)

	n.EthClient_1, n.EthClient_2, err = getEthClients()
	if err != nil {
		return
	}

	key_file := "keys/" + name + "-ecdsa.key"
	bls_key_file := "keys/" + name + "-bn256.key"

	blsAddr, err := common2.BLSAddrFromKeyFile(bls_key_file)
	if err != nil {
		return
	}

	q, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	if err != nil {
		return
	}

	words := strings.Split(string(q.P2pAddress), "/")

	registeredPort, err := strconv.Atoi(words[4])
	if err != nil {
		logrus.Fatalf("Can't obtain port %d, %v", registeredPort, err)
	}
	logrus.Infof("PORT %d", registeredPort)

	registeredAddress := words[2]

	registeredPeer, err := ioutil.ReadFile("keys/" + name + "-peer.env")
	if err != nil {
		logrus.Fatalf("File %s reading error: %v", "keys/"+name+"-peer.env", err)
	}
	logrus.Infof("Node address: %s nodeAddress from contract: %s", string(registeredPeer), string(q.P2pAddress))

	if string(registeredPeer) != string(q.P2pAddress) {
		logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", string(registeredPeer), string(q.P2pAddress))
	}
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, key_file, registeredPort, registeredAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("host id %v address %v", n.Host.ID(), n.Host.Addrs()[0])
	n.DiscoveryPeers = make(addrList, 0)
	n.DiscoveryPeers, err = n.AddActiveDiscoveryPeersFromContract(n.DiscoveryPeers)
	if err != nil {
		logrus.Fatal(err)
	}
	n.Dht, err = n.initDHT()
	if err != nil {
		return
	}
	logrus.Infof("DiscoveryPeers %d", len(n.DiscoveryPeers))

	n.Discovery = discovery.NewRoutingDiscovery(n.Dht)
	_ = n.ListenAndAddNodeToPeersFromEvent()
	//go n.InitNodeDiscoveryPeers()

	n.PrivKey, n.BLSAddress, err = n.KeysFromFilesByConfigName(name)
	if err != nil {
		return
	}
	_, err = n.ListenNodeOracleRequest()
	if err != nil {
		logrus.Errorf(err.Error())
		panic(err)
	}

	n.ListenReceiveRequest(n.EthClient_2, common.HexToAddress(config.Config.PROXY_NETWORK2))

	logrus.Info("bridge started")
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	//n.Server.Start(port)

	run(n.Host, cancel)
	return
}

func getEthClients() (c1 *ethclient.Client, c2 *ethclient.Client, err error) {
	logrus.Tracef("config.Config.NETWORK_RPC_1 %s", config.Config.NETWORK_RPC_1)
	c1, err = ethclient.Dial(config.Config.NETWORK_RPC_1)
	if err != nil {
		return
	}

	c2, err = ethclient.Dial(config.Config.NETWORK_RPC_2)
	if err != nil {
		return
	}
	return
}

func (n Node) AddActiveDiscoveryPeersFromContract(peers addrList) (peers2 addrList, err error) {
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	logrus.Infof("found %d peers in contract", len(nodes))
	for _, node := range nodes {
		peerMA, err := multiaddr.NewMultiaddr(string(node.P2pAddress[:]))
		if err != nil {
			panic(err)
		}
		peers = append(peers, peerMA)
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerMA)

		var wg sync.WaitGroup
		if n.Host.ID().Pretty() != peerinfo.ID.Pretty() && n.Host.Network().Connectedness(peerinfo.ID) != network.Connected {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := n.Host.Connect(n.Ctx, *peerinfo); err != nil {
					logrus.Tracef("Error while connecting to node %q: %-v\n not going to add", peerinfo, err)

				} else {
					logrus.Infof("Adding to DiscoveryPeers connected node: %q\n", peerinfo)

				}
			}()
		}
	}
	peers2 = peers
	return
}

func (n Node) AddPeer(peerAddr string, peers addrList) (peers2 addrList, err error) {
	peerMA, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		panic(err)
	}
	peers = append(peers, peerMA)
	peerinfo, _ := peer.AddrInfoFromP2pAddr(peerMA)
	var wg sync.WaitGroup
	if n.Host.ID().Pretty() != peerinfo.ID.Pretty() && n.Host.Network().Connectedness(peerinfo.ID) != network.Connected {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.Host.Connect(n.Ctx, *peerinfo); err != nil {
				logrus.Tracef("Error while connecting to node %q: %-v\n not going to add", peerinfo, err)

			} else {
				logrus.Infof("Adding to DiscoveryPeers connected node: %q\n", peerinfo)

			}
		}()
	}

	peers2 = peers
	return
}

/*func(n Node) InitNodeDiscoveryPeers() {
	for _, addr := range n.DiscoveryPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(addr)
		logrus.Trace("addr", addr.String())
		logrus.Tracef("self ID %s host ID %s", n.Host.ID().Pretty(), peerinfo.ID.Pretty())
		var wg sync.WaitGroup
		for _, peerAddr := range n.DiscoveryPeers {
			peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
			if n.Host.ID().Pretty() != peerinfo.ID.Pretty() && n.Host.Network().Connectedness(peerinfo.ID) != network.Connected {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := n.Host.Connect(n.Ctx, *peerinfo); err != nil {
						logrus.Errorf("InitNodeDiscoveryPeers Error while connecting to node %q: %-v", peerinfo, err)
					} else {
						logrus.Tracef("InitNodeDiscoveryPeers Connection established with node: %q", peerinfo)
					}
				}()
			}
		}
		wg.Wait()

		//if n.Host.ID().Pretty() != peerinfo.ID.Pretty() && n.Host.Network().Connectedness(peerinfo.ID) != network.Connected {
		//	_, err := n.Host.Network().DialPeer(n.Ctx, peerinfo.ID)
		//	if err != nil {
		//		logrus.Errorf("InitNodeDiscoveryPeers Host.Network().Connectedness PROBLEM: %v", err)
		//		continue
		//	 } else {
		//		logrus.Infof("InitNodeDiscoveryPeers Connected to peer %s", peerinfo.ID)
		//	}
		//}



	}
}
*/

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {
	dht, err = libp2p.NewDHT(n.Ctx, n.Host, n.DiscoveryPeers)
	if err != nil {
		return
	}
	err = dht.Bootstrap(n.Ctx)
	if err != nil {
		return
	}
	return
}
