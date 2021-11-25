package consensus

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/sirupsen/logrus"
	mrand "math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

type Protocol struct {
	pubsub        *pubsub.PubSub // PubSub of each individual node
	mainTopic     string
	topics        map[string]*pubsub.Topic        // PubSub topics
	subscriptions map[string]*pubsub.Subscription // Subscription of individual node
	mx            *sync.Mutex
}

func (c *Protocol) ListPeersByTopic(topicName string) []peer.ID {
	c.mx.Lock()
	defer c.mx.Unlock()
	if topic, ok := c.topics[topicName]; ok {

		return topic.ListPeers()
	}

	return []peer.ID{}
}

// Broadcast Uses PubSub publish to broadcast messages to other peers
func (c *Protocol) Broadcast(msgBytes []byte) {
	// Broadcasting to a topic in PubSub
	go func(msgBytes []byte, topic *pubsub.Topic) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := topic.Publish(ctx, msgBytes)
		if err != nil {
			err = fmt.Errorf("publish to main topic error: %w", err)
			if errors.Is(err, pubsub.ErrTopicClosed) {
				logrus.Warn(err)
			} else {
				logrus.Error(err)
			}
			return
		}

	}(msgBytes, c.getTopic(c.mainTopic))
}

func (c *Protocol) BroadcastTo(topic *pubsub.Topic, msgBytes []byte) {
	// Broadcasting to a topic in PubSub
	go func(msgBytes []byte, topic *pubsub.Topic) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := topic.Publish(ctx, msgBytes)
		if err != nil {
			err = fmt.Errorf("broadcast to topic %s error: %w", topic.String(), err)
			if errors.Is(err, pubsub.ErrTopicClosed) {
				logrus.Warn(err)
			} else {
				logrus.Error(err)
			}
			return
		}

	}(msgBytes, topic)
}

func (c *Protocol) getTopic(topicName string) *pubsub.Topic {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.topics[topicName]
}

func (c *Protocol) JoinTopic(topicName string) (topic *pubsub.Topic, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if topic, ok := c.topics[topicName]; ok {

		return topic, nil
	} else if topic, err := c.pubsub.Join(topicName); err != nil {

		return nil, err
	} else {
		c.topics[topicName] = topic

		return topic, nil
	}
}

func (c *Protocol) LeaveTopic(topicName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	if topic, ok := c.topics[topicName]; ok && topicName != c.mainTopic {
		if subscription, ok := c.subscriptions[topicName]; ok {
			subscription.Cancel()
			delete(c.subscriptions, topicName)
		}
		delete(c.topics, topicName)
		return topic.Close()
	}
	return nil
}

// Receive gets message from main topic PubSub in a blocking way
func (c *Protocol) Receive(ctx context.Context) *[]byte {

	c.mx.Lock()
	if subscription, ok := c.subscriptions[c.mainTopic]; ok {
		c.mx.Unlock()
		// Blocking function for consuming newly received messages
		// We can access message here
		msg, err := subscription.Next(ctx)
		// handling canceled subscriptions
		if err != nil {
			return nil
		}

		msgBytes := msg.Data

		return &msgBytes
	} else {
		c.mx.Unlock()
		return nil
	}
}

// ReceiveFrom gets message from PubSub in a blocking way for topic
func (c *Protocol) ReceiveFrom(topicName string, ctx context.Context) *[]byte {
	c.mx.Lock()
	if subscription, ok := c.subscriptions[topicName]; ok {
		c.mx.Unlock()
		// Blocking function for consuming newly received messages
		// We can access message here
		msg, err := subscription.Next(ctx)
		// handling canceled subscriptions
		if err != nil {
			return nil
		}

		msgBytes := msg.Data

		return &msgBytes
	} else {
		c.mx.Unlock()
		logrus.Warnf("subscription for topic %s not found", topicName)
		return nil
	}
}

func (c *Protocol) Disconnect() {
	c.mx.Lock()
	defer c.mx.Unlock()
	for topicName, topic := range c.topics {
		if subscription, ok := c.subscriptions[topicName]; ok {
			subscription.Cancel()
			delete(c.subscriptions, topicName)
		}
		if err := c.pubsub.UnregisterTopicValidator(topicName); err != nil {

			logrus.Error(err)
		}
		if err := topic.Close(); err != nil {

			logrus.Error(err)
		}
		delete(c.topics, topicName)
	}
}

func (c *Protocol) Subscribe(topicName string) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	if topic, ok := c.topics[topicName]; ok {
		if sub, subscribed := c.subscriptions[topicName]; subscribed && sub != nil {

			return nil
		} else if subscription, err := topic.Subscribe(); err != nil {

			return fmt.Errorf("topic %s subscribe error %w", topicName, err)
		} else {
			c.subscriptions[topicName] = subscription

			return nil
		}
	} else {

		return fmt.Errorf("topic %s not joined", topicName)
	}
}

// CreatePeer creates a peer on localhost and configures it to use libp2p.
func (c *Protocol) CreatePeer(nodeId int, port int) *core.Host {
	// Creating a node
	h, err := createHost(port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node %v is %s\n", nodeId, GetLocalhostAddress(h))

	return &h
}

// CreatePeerWithIp creates a peer on specified ip and port and configures it to use libp2p.
func (c *Protocol) CreatePeerWithIp(nodeId int, ip string, port int) *core.Host {
	// Creating a node
	h, err := createHostWithIp(nodeId, ip, port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node %v is %s\n", nodeId, GetLocalhostAddress(h))

	return &h
}

// NewPubSubWithTopic creates a PubSub for the peer and also subscribes to a main topic
func NewPubSubWithTopic(h core.Host, dht *dht.IpfsDHT, mainTopicName string) (*Protocol, error) {

	optsPS := []pubsub.Option{
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign | 1),
		pubsub.WithPeerOutboundQueueSize(512),
		pubsub.WithValidateWorkers(runtime.NumCPU() * 2),
		// pubsub.WithEventTracer(NewTracer()),
	}

	if pubSub, err := pubsub.NewGossipSub(context.Background(), h, optsPS...); err != nil {
		logrus.Error(fmt.Errorf("init gosip error: %w", err))

		return nil, err
	} else if topic, err := pubSub.Join(mainTopicName); err != nil {
		logrus.Error(fmt.Errorf("join main topic %s error: %w", mainTopicName, err))

		return nil, err
	} else if subscription, err := topic.Subscribe(); err != nil {
		logrus.Error(fmt.Errorf("subscribe to main topic: %s error: %w", mainTopicName, err))

		return nil, err
	} else {

		c := &Protocol{
			pubsub:        pubSub,
			mainTopic:     mainTopicName,
			topics:        make(map[string]*pubsub.Topic, 20),
			subscriptions: make(map[string]*pubsub.Subscription, 20),
			mx:            new(sync.Mutex),
		}

		// if topic, err := pubSub.Join(mainTopicName); err != nil {
		//     logrus.Error(fmt.Errorf("join main topic %s error: %w", mainTopicName, err))
		//
		//     return nil, err
		// } else if subscription, err := pubSub.Subscribe(mainTopicName); err != nil {
		//     logrus.Error(fmt.Errorf("subscribe to main topic: %s error: %w", mainTopicName, err))
		//
		//     return nil, err
		// } else {
		c.subscriptions[mainTopicName] = subscription
		c.topics[mainTopicName] = topic
		// }
		// time.Sleep(10 * time.Second)

		return c, nil
	}
}

// // InitializePubSubWithTopicAndPeers creates a PubSub for the peer with some extra parameters
// func (c *Protocol) InitializePubSubWithTopicAndPeers(h core.Host, topic string, peerAddrs []peer.AddrInfo) {
// 	var err error
// 	// Creating pubsub
// 	// every peer has its own PubSub
// 	c.pubsub, err = applyPubSubDeatiled(h, peerAddrs)
// 	if err != nil {
// 		fmt.Printf("Error : %v\n", err)
// 		return
// 	}
//
// 	c.InitializePubSubWithTopic(h, topic)
//
// 	// Creating a subscription and subscribing to the topic
// 	c.subscription, err = c.pubsub.Subscribe(c.topic.String())
// 	if err != nil {
// 		fmt.Printf("Error : %v\n", err)
// 		return
// 	}
//
// }

/*func (c *Protocol) SetTopic(topic string) {
	c.topic = topic
}
*/
// createHost creates a host with some default options and a signing identity
func createHost(port int) (core.Host, error) {
	// Producing private key
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	// Starting a peer with default configs
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(sk),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// createHostWithIp creates a host with some default options and a signing identity on specified port and ip
func createHostWithIp(nodeId int, ip string, port int) (core.Host, error) {
	// Producing private key using nodeId
	r := mrand.New(mrand.NewSource(int64(nodeId)))

	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	// Starting a peer with default configs
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", strconv.Itoa(port))),
		libp2p.Identity(sk),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// createHostQUIC creates a host with QUIC as transport layer implementation
// func createHostQUIC(port int) (core.Host, error) {
// 	// Producing private key
// 	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
//
// 	quicTransport, err := quic.NewTransport(priv, nil, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Starting a peer with QUIC transport
// 	opts := []libp2p.Option{
// 		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
// 		libp2p.Transport(quicTransport),
// 		libp2p.Identity(priv),
// 		libp2p.DefaultMuxers,
// 		libp2p.DefaultSecurity,
// 	}
//
// 	h, err := libp2p.New(context.Background(), opts...)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return h, nil
// }

// createHostWebSocket creates a host with WebSocket as transport layer implementation
// func createHostWebSocket(port int) (core.Host, error) {
//
// 	// Starting a peer with QUIC transport
// 	opts := []libp2p.Option{
// 		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
// 		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", port)),
// 		libp2p.Transport(ws.New),
// 		libp2p.DefaultMuxers,
// 		libp2p.DefaultSecurity,
// 	}
//
// 	h, err := libp2p.New(context.Background(), opts...)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return h, nil
// }

// GetLocalhostAddress is used for getting address of hosts
func GetLocalhostAddress(h core.Host) string {
	for _, addr := range h.Addrs() {
		if strings.Contains(addr.String(), "127.0.0.1") {
			return addr.String() + "/p2p/" + h.ID().Pretty()
		}
	}

	return ""
}

// applyPubSub creates a new GossipSub with message signing
// func applyPubSub(h core.Host) (*pubsub.PubSub, error) {
// 	optsPS := []pubsub.Option{
// 		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
// 	}
//
// 	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
// }

// applyPubSubDetailed creates a new GossipSub with message signing
// func applyPubSubDeatiled(h core.Host, addrInfos []peer.AddrInfo) (*pubsub.PubSub, error) {
// 	optsPS := []pubsub.Option{
// 		// pubsub.WithMessageSigning(true),
// 		pubsub.WithPeerExchange(true),
// 		// pubsub.WithMessageIdFn(func(pmsg *pubsubpb.Message) string {
// 		//	hash := blake2b.Sum256(pmsg.Data)
// 		//	return string(hash[:])
// 		// }),
// 		pubsub.WithDirectPeers(addrInfos),
// 		pubsub.WithFloodPublish(true),
// 		pubsub.WithDirectConnectTicks(7),
// 	}
// 	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
// }

// connectHostToPeer is used for connecting a host to another peer
// func connectHostToPeer(h core.Host, connectToAddress string) {
// 	// Creating multi address
// 	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
// 	if err != nil {
// 		fmt.Printf("Error : %v\n", err)
// 		return
// 	}
//
// 	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
// 	if err != nil {
// 		fmt.Printf("Error : %v\n", err)
// 		return
// 	}
//
// 	err = h.Connect(context.Background(), *pInfo)
// 	if err != nil {
// 		fmt.Printf("Error : %v\n", err)
// 		return
// 	}
// }

func ConnectHostToPeerWithError(h core.Host, connectToAddress string) (err error) {
	// Creating multi address
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}
	return
}

func (c *Protocol) MainTopic() *pubsub.Topic {

	return c.getTopic(c.mainTopic)
}

func (c *Protocol) Topic(topicName string) *pubsub.Topic {

	return c.getTopic(topicName)
}
