package node

import (
	"context"
	"errors"
	common2 "github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/DigiU-Lab/p2p-bridge/config"
	"github.com/DigiU-Lab/p2p-bridge/helpers"
	"github.com/DigiU-Lab/p2p-bridge/libp2p"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode"
)

func loadNodeConfig(path string) (err error) {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
		return
	}
	logrus.Printf("started in directory %s", dir)
	err = config.LoadConfigAndArgs(path)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	return
}

func NodeInit(path, name string) (err error) {
	logrus.Print("nodeInit START")
	hostName, _ := os.Hostname()
	keysList := os.Getenv("ECDSA_KEY_1")
	strNum := strings.TrimPrefix(hostName, "p2p-bridge_node_")
	strNum = strings.TrimRightFunc(strNum, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	nodeHostId, _ := strconv.Atoi(strNum)
	if nodeHostId == 0 || nodeHostId > len(keysList)-1 {
		nodeHostId = rand.Intn(len(keysList))
	}

	config.Config.ECDSA_KEY_1 = strings.Split(keysList, ",")[nodeHostId-1]

	keysList2 := os.Getenv("ECDSA_KEY_2")
	config.Config.ECDSA_KEY_2 = strings.Split(keysList2, ",")[nodeHostId-1]

	logrus.Printf("----------> %v key1 %v key2 %v", string(hostName[len(hostName)-1]), config.Config.ECDSA_KEY_1, config.Config.ECDSA_KEY_2)

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	blsAddr, pub, err := common2.CreateBN256Key(name)
	if err != nil {
		return
	}

	logrus.Printf("pubkey %v", pub)

	err = common2.GenECDSAKey(name)
	if err != nil {
		return
	}

	logrus.Printf("keyfile %v", "keys/"+name+"-ecdsa.key")
	h, err := libp2p.NewHostFromKeyFila(context.Background(), "keys/"+name+"-ecdsa.key", 0)
	if err != nil {
		return
	}
	nodeURL := libp2p.WriteHostAddrToConfig(h, "keys/"+name+"-peer.env")
	c1, c2, err := getEthClients()

	if err != nil {
		return
	}

	logrus.Printf("nodelist 1 %v blsAddress %v", common.HexToAddress(config.Config.ECDSA_KEY_1), pub)
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

	logrus.Printf("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func NewNode(path, name string, port int) (err error) {

	err = loadNodeConfig(path)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	n := &Node{
		Ctx:               ctx,
		CurrentRendezvous: "Init",
	}

	n.pKey, err = common2.ToECDSAFromHex(config.Config.ECDSA_KEY_1)
	if err != nil {
		return
	}

	server := n.NewBridge()
	n.Server = *server
	logrus.Printf("n.Config.PORT_1 %d", config.Config.PORT_1)

	n.EthClient_1, n.EthClient_2, err = getEthClients()
	if err != nil {
		return
	}

	/*	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
		if err != nil {
			return
		}
		logrus.Printf("------------> node count %v", len(nodes))
		for _, node := range nodes {
			logrus.Printf("NODES: %v", string(node.P2pAddress[:]))
		}
	*/

	key_file := "keys/" + name + "-ecdsa.key"
	bls_key_file := "keys/" + name + "-bn256.key"
	blsAddr := common2.BLSAddrFromKeyFile(bls_key_file)
	q, err := common2.GetNode(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1), blsAddr)
	if err != nil {
		return
	}

	words := strings.Split(string(q.P2pAddress), "/")

	registeredPort, err := strconv.Atoi(words[4])
	if err == nil {
		logrus.Error("Caanot obtain port %d, %v", registeredPort, err)
	}
	logrus.Printf(" <<<<<<<<<<<<<<<<<<<<<<< PORT %d >>>>>>>>>>>>>>>>>>>>>>>>>>>>", registeredPort)
	n.Host, err = libp2p.NewHostFromKeyFila(n.Ctx, key_file, registeredPort)
	if err != nil {
		return
	}
	logrus.Printf(" <<<<<<<<<<<<<<<<<<<<<<< HOST ID %v >>>>>>>>>>>>>>>>>>>>>>>>>>>>", n.Host.ID())
	n.Dht, err = n.initDHT()

	n.P2PPubSub = n.initNewPubSub()

	n.NodeBLS, err = n.NewBLSNode(path, name)
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}
	if n.NodeBLS == nil {
		err = errors.New("newBLSNode NIL")
		logrus.Errorf(err.Error())
	} else {
		n.Dht, err = n.initDHT()
		if err != nil {
			logrus.Errorf(err.Error())
			return err
		}
		logrus.Printf("---------- ListenNodeOracleRequest -------------")
		_, err = n.ListenNodeOracleRequest()
		if err != nil {
			logrus.Errorf(err.Error())
		}
		logrus.Printf("---------- ListenReceiveRequest -------------")
		helpers.ListenReceiveRequest(n.EthClient_2, common.HexToAddress(config.Config.PROXY_NETWORK2))

	}
	logrus.Print("newBLSNode STARTED /////////////////////////////////")
	/*err = n.runRPCService()
	if err != nil {
		return
	}*/

	if port == 0 {
		port = config.Config.PORT_1
	}
	n.Server.Start(port)

	run(n.Host, cancel)
	return
}

func getEthClients() (c1 *ethclient.Client, c2 *ethclient.Client, err error) {
	logrus.Printf("config.Config.NETWORK_RPC_1 %s", config.Config.NETWORK_RPC_1)
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

func (n Node) initDHT() (dht *dht.IpfsDHT, err error) {
	var bootstrapPeers []multiaddr.Multiaddr
	nodes, err := common2.GetNodesFromContract(n.EthClient_1, common.HexToAddress(config.Config.NODELIST_NETWORK1))
	if err != nil {
		return
	}
	for _, node := range nodes {
		//logrus.Printf("Node:%d\n %v\n %v\n %v\n %v\n", i, node.Enable, node.NodeWallet, string(node.P2pAddress[:]), string(node.BlsPubKey[:]))
		peerMA, err := multiaddr.NewMultiaddr(string(node.P2pAddress[:]))
		if err != nil {
			return nil, err
		}
		bootstrapPeers = append(bootstrapPeers, peerMA)
	}
	return libp2p.NewDHT(n.Ctx, n.Host, bootstrapPeers)
}
