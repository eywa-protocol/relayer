package consensus

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	ws "github.com/libp2p/go-ws-transport"
	"github.com/multiformats/go-multiaddr"
)

type Protocol struct {
	pubsub       *pubsub.PubSub       // PubSub of each individual node
	subscription *pubsub.Subscription // Subscription of individual node
	topic        *pubsub.Topic        // PubSub topic
}

func (c *Protocol) ListPeersByTopic(topic string) []peer.ID {
	return c.pubsub.ListPeers(topic)
}

// Broadcast Uses PubSub publish to broadcast messages to other peers
func (c *Protocol) Broadcast(msgBytes []byte) {
	// Broadcasting to a topic in PubSub
	go func(msgBytes []byte, topic string, pubsub *pubsub.PubSub) {

		err := pubsub.Publish(topic, msgBytes)
		if err != nil {
			fmt.Printf("Publish Error: %v\n", err)
			return
		}

	}(msgBytes, c.topic.String(), c.pubsub)
}

func (c *Protocol) JoinTopic(topicMsg string) (topic *pubsub.Topic, err error) {
	return c.pubsub.Join(topicMsg)
}

// Send uses Broadcast for sending messages
func (c *Protocol) Send(msgBytes []byte, id int) {
	// In libp2p implementation, we also broadcast instead of sending directly. So Acks will be broadcast in this case.
	c.Broadcast(msgBytes)
}

// Receive gets message from PubSub in a blocking way
func (c *Protocol) Receive(ctx context.Context) *[]byte {

	// Blocking function for consuming newly received messages
	// We can access message here
	msg, err := c.subscription.Next(ctx)
	// handling canceled subscriptions
	if err != nil {
		return nil
	}

	msgBytes := msg.Data

	return &msgBytes
}

func (c *Protocol) Disconnect() {
	c.subscription.Cancel()
	c.pubsub.UnregisterTopicValidator(c.topic.String())
}

func (c *Protocol) DisconnectTopic(topic string) {
	c.pubsub.UnregisterTopicValidator(topic)

}

func (c *Protocol) Reconnect(topic string) {
	var err error
	//if topic != "" {
	//	c.topic = topic
	//}

	c.subscription, err = c.pubsub.Subscribe(c.topic.String())
	if err != nil {
		panic(err)
	}
	fmt.Println("RECONNECT to topic ", c.topic)
}

// Cancel unsubscribes a node from pubsub
func (c *Protocol) Cancel(cancelTime int, reconnectTime int) {
	go func() {
		time.Sleep(time.Duration(cancelTime) * time.Millisecond)
		fmt.Println("	CANCELING	")
		c.subscription.Cancel()
		time.Sleep(time.Duration(reconnectTime) * time.Millisecond)
		fmt.Println("	RESUBBING	")
		c.subscription, _ = c.pubsub.Subscribe(c.topic.String())
	}()
}

// createPeer creates a peer on localhost and configures it to use libp2p.
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

// initializePubSub creates a PubSub for the peer and also subscribes to a topic
func (c *Protocol) InitializePubSubWithTopic(h core.Host, topic string) {
	var err error
	// Creating pubsub
	// every peer has its own PubSub
	c.pubsub, err = applyPubSub(h)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	c.topic, err = c.pubsub.Join(topic)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	// Creating a subscription and subscribing to the topic
	c.subscription, err = c.pubsub.Subscribe(c.topic.String())
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

}

// InitializePubSubWithTopicAndPeers creates a PubSub for the peer with some extra parameters
func (c *Protocol) InitializePubSubWithTopicAndPeers(h core.Host, topic string, peerAddrs []peer.AddrInfo) {
	var err error
	// Creating pubsub
	// every peer has its own PubSub
	c.pubsub, err = applyPubSubDeatiled(h, peerAddrs)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	c.InitializePubSubWithTopic(h, topic)

	// Creating a subscription and subscribing to the topic
	c.subscription, err = c.pubsub.Subscribe(c.topic.String())
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

}

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
func createHostQUIC(port int) (core.Host, error) {
	// Producing private key
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	quicTransport, err := quic.NewTransport(priv, nil, nil)
	if err != nil {
		return nil, err
	}

	// Starting a peer with QUIC transport
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
		libp2p.Transport(quicTransport),
		libp2p.Identity(priv),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// createHostWebSocket creates a host with WebSocket as transport layer implementation
func createHostWebSocket(port int) (core.Host, error) {

	// Starting a peer with QUIC transport
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", port)),
		libp2p.Transport(ws.New),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

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
func applyPubSub(h core.Host) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}

// applyPubSubDetailed creates a new GossipSub with message signing
func applyPubSubDeatiled(h core.Host, addrInfos []peer.AddrInfo) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
		pubsub.WithPeerExchange(true),
		//pubsub.WithMessageIdFn(func(pmsg *pubsubpb.Message) string {
		//	hash := blake2b.Sum256(pmsg.Data)
		//	return string(hash[:])
		//}),
		pubsub.WithDirectPeers(addrInfos),
		pubsub.WithFloodPublish(true),
		pubsub.WithDirectConnectTicks(7),
	}
	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}

// connectHostToPeer is used for connecting a host to another peer
func connectHostToPeer(h core.Host, connectToAddress string) {
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
}

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

func (c *Protocol) UnregisterTopicValidator() {
	c.pubsub.UnregisterTopicValidator(c.topic.String())
}
