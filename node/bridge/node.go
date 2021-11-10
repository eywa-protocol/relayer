package bridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/libp2p/go-flow-metrics"
	bls_consensus "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/model"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/forward"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/network"
	discovery "github.com/libp2p/go-libp2p-discovery"
	pubSub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	messageSigPb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/protobuf/message"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/helpers"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

var ErrContextDone = errors.New("interrupt on context done")

const minConsensusNodesCount = 5

type EpochKeys struct {
	Id             int             // Id of the node, starting from 0
	EpochPublicKey bls.PublicKey   // Arrtegated public key of all participants of the current epoch
	MembershipKey  bls.Signature   // Membership key of this node
	PublicKeys     []bls.PublicKey // Public keys of all nodes
}

type Node struct {
	base.Node
	cMx            *sync.Mutex
	Clients        map[string]Client
	nonceMx        *sync.Mutex
	P2PPubSub      *bls_consensus.Protocol
	signerKey      *ecdsa.PrivateKey
	PrivKey        bls.PrivateKey
	uptimeRegistry *flow.MeterRegistry
	gsnClient      *gsn.Client
	session        *model.Node
	BlsNodeId      int
	EpochKeys
}

func (n Node) StartProtocolByOracleRequest(event *wrappers.BridgeOracleRequest, wg *sync.WaitGroup) {
	defer n.session.DisconnectPubSub()
	defer n.session.Topic.Close()
	defer wg.Done()

	consensusChannel := make(chan bool)
	n.session.AdvanceStep(0)
	time.Sleep(time.Second)
	wg.Add(1)

	go n.session.WaitForProtocolMsg(consensusChannel, wg)
	select {
	case <-consensusChannel:
		logrus.Infof("LEADER:%s == MYMNODE:%s", n.session.Leader.Pretty(), n.Host.ID().Pretty())
		if n.session.Leader.Pretty() == n.Host.ID().Pretty() {
			logrus.Info("LEADER going to Call external chain contract method")
			_, err := n.ReceiveRequestV2(event)
			if err != nil {
				logrus.Error(fmt.Errorf("%w", err))
			}
		}
	case <-time.After(5000 * time.Millisecond):
		logrus.Warn("WaitGroup timed out..")

	}
	if err := n.session.Topic.Close(); err != nil {
		logrus.Warnf("nodeBls.Topic.Close: %v", err)
	}
	logrus.Info("CLOSED TOPIC")
	logrus.Info("The END of Protocol")
}

func (n Node) nodeExists(client Client, nodeIdAddress common.Address) bool {
	node, err := common2.GetNode(client.EthClient, client.ChainCfg.NodeRegistryAddress, nodeIdAddress)
	if err != nil || node.Owner == common.HexToAddress("0") {
		return false
	}
	return true

}

func (n Node) getNodeBlsId() (id int) {
	// TODO put here service chaun client !!!
	client := n.Clients[config.Bridge.Chains[0].ChainId.String()]
	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	if !n.nodeExists(client, nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())
		return -1
	} else {
		logrus.Tracef("Host.ID() %v ", n.Host.ID())
		node, err := client.NodeRegistry.GetNode(nodeIdAddress)
		if err != nil {
			logrus.Errorf("NodeRegistry.GetNode error: %x", err)
			return -1
		}
		id = int(node.NodeId.Int64())

	}
	return
}

func GetPubKeysFromContract(client Client) (publicKeys []bls.PublicKey, err error) {
	publicKeys = make([]bls.PublicKey, 0)
	nodes, err := common2.GetNodesFromContract(client.EthClient, client.ChainCfg.NodeRegistryAddress)
	if err != nil {
		return
	}
	for _, node := range nodes {
		p, err := bls.UnmarshalPublicKey([]byte(node.BlsPubKey))
		if err != nil {
			panic(err)
		}
		publicKeys = append(publicKeys, p)
	}
	return
}

func (n Node) KeysFromFilesByConfigName(name string) (prvKey bls.PrivateKey, err error) {

	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err = common2.ReadBlsPrivateKeyFromFile(nodeKeyFile)
	if err != nil {
		return
	}

	return
}

func (n Node) NewBLSSession(topic *pubSub.Topic, publicKeys []bls.PublicKey, topicParticipants []peer.ID) (blsNode *model.Node) {
	return &model.Node{
		Ctx:            n.Ctx,
		Id:             n.BlsNodeId,
		TimeStep:       0,
		ThresholdWit:   len(topicParticipants)/2 + 1,
		ThresholdAck:   len(topicParticipants)/2 + 1,
		Acks:           0,
		ConvertMsg:     &messageSigPb.Convert{},
		Comm:           n.P2PPubSub,
		History:        make([]model.MessageWithSig, 0),
		SigMask:        bls.EmptyMultisigMask(),
		PublicKeys:     publicKeys,
		PrivateKey:     n.PrivKey,
		EpochPublicKey: n.EpochPublicKey,
		MembershipKey:  n.MembershipKey,
		PartPublicKey:  bls.ZeroPublicKey(),
		PartSignature:  bls.ZeroSignature(),
		Leader:         "",
		Topic:          topic,
	}
}

/*func (n Node) NewBLSNode(topic *pubSub.Topic, client Client) (blsNode *model.Node, err error) {
	publicKeys, err := GetPubKeysFromContract(client)
	if err != nil {
		return
	}

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	if !n.nodeExists(client, nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())

	} else {
		logrus.Tracef("Host.ID() %v ", n.Host.ID())
		node, err := client.NodeRegistry.GetNode(nodeIdAddress)
		if err != nil {
			return nil, err
		}

		blsNode = func() *model.Node {
			ctx, cancel := context.WithDeadline(n.Ctx, time.Now().Add(10*time.Second))
			defer cancel()
			for {
				topicParticipants := topic.ListPeers()
				topicParticipants = append(topicParticipants, n.Host.ID())
				logrus.Tracef("len(topicParticipants) = [ %d ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.Dht.RoutingTable().ListPeers()))
				if len(topicParticipants) > minConsensusNodesCount && len(topicParticipants) > len(n.P2PPubSub.ListPeersByTopic(config.Bridge.Rendezvous))/2+1 {
					blsNode = &model.Node{
						Ctx:               n.Ctx,
						Id:                int(node.NodeId.Int64()),
						TimeStep:          0,
						ThresholdWit:      len(topicParticipants)/2 + 1,
						ThresholdAck:      len(topicParticipants)/2 + 1,
						Acks:              0,
						ConvertMsg:        &messageSigPb.Convert{},
						Comm:              n.P2PPubSub,
						History:           make([]model.MessageWithSig, 0),
						SigMask:           bls.EmptyMultisigMask(),
						PublicKeys:        publicKeys,
						PrivateKey:        n.PrivKey,
						EpochPublicKey:    n.EpochPublicKey,
						MembershipKey:     n.MembershipKey,
						PartPublicKey:     bls.ZeroPublicKey(),
						PartSignature:     bls.ZeroSignature(),
						Participants:      topicParticipants,
						Leader:            "",
						Topic: topic,
					}
					break
				}
				if ctx.Err() != nil {
					logrus.Warnf("Not enaugh participants %d , %v", len(topicParticipants), ctx.Err())
					_ = topic.Close()
					break
				}
				time.Sleep(300 * time.Millisecond)
			}

			return blsNode
		}()

	}
	return
}*/

func (n *Node) ListenReceiveRequest(clientNetwork *ethclient.Client, proxyNetwork common.Address) {

	bridgeFilterer, err := wrappers.NewBridge(proxyNetwork, clientNetwork)
	if err != nil {
		return
	}
	channel := make(chan *wrappers.BridgeReceiveRequest)
	opt := &bind.WatchOpts{}

	sub, err := bridgeFilterer.WatchReceiveRequest(opt, channel)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case _ = <-sub.Err():
				break
			case e := <-channel:
				logrus.Infof("ReceiveRequest: %v %v", e.ReqId, e.ReceiveSide)
				// TODO disconnect from topic
				/** TODO:
				Is transaction true, otherwise repeat to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}

func (n *Node) SetSessionForTopic(topic *pubSub.Topic) {

}

func (n *Node) ListenNodeOracleRequest(channel chan *wrappers.BridgeOracleRequest, wg *sync.WaitGroup, chainId *big.Int) (err error) {
	opt := &bind.WatchOpts{}
	client, _, err := n.GetNodeClientOrRecreate(chainId)
	if err != nil {
		return err
	}
	sub, err := client.BridgeFilterer.WatchOracleRequest(opt, channel)
	if err != nil {
		logrus.Errorf("WatchOracleRequest can't %v", err)
		return
	}

	checkClientTimer := time.NewTicker(10 * time.Second)
	recreateOnTimer := false
	mx := new(sync.Mutex)
	wg.Add(1)
	go func(subPtr *event.Subscription, clientPtr *Client) {
		defer func() {
			checkClientTimer.Stop()
			wg.Done()
		}()
	reqLoop:
		for {
			select {
			case <-checkClientTimer.C:
				if err := func() error {
					mx.Lock()
					defer mx.Unlock()
					if recreateOnTimer {
						clientRecreated := false
						client, clientRecreated, err = n.GetNodeClientOrRecreate(chainId)
						if err != nil {
							logrus.WithField(field.CainId, chainId.String()).
								Error(fmt.Errorf("can not get client for network [%s] on error:%w",
									chainId.String(), err))
							time.Sleep(1 * time.Second)
							return err
						}
						if clientRecreated {
							logrus.Info("resubscribe recreated client on timer")
							sub, err = client.BridgeFilterer.WatchOracleRequest(opt, channel)
							if err != nil {
								logrus.Error(fmt.Errorf("WatchOracleRequest can't %w", err))
								time.Sleep(1 * time.Second)
								return err
							}
						}
					}
					return nil
				}(); err != nil {
					continue
				}
			case err := <-(*subPtr).Err():
				if err != nil {
					logrus.Error(fmt.Errorf("OracleRequest subscription error: %w", err))
					clientRecreated := false
					client, clientRecreated, err = n.GetNodeClientOrRecreate(chainId)
					if err != nil {
						logrus.WithField(field.CainId, chainId.String()).
							Error(fmt.Errorf("can not get client for network [%s] on error:%w",
								chainId.String(), err))
						time.Sleep(1 * time.Second)
						mx.Lock()
						recreateOnTimer = true
						mx.Unlock()
						continue
					}
					if clientRecreated {
						logrus.Infof("subscribe to OracleRequest on recreated client on sub err: %v", err)
						sub, err = client.BridgeFilterer.WatchOracleRequest(opt, channel)
						if err != nil {
							logrus.Error(fmt.Errorf("WatchOracleRequest can't %w", err))
							time.Sleep(1 * time.Second)
							mx.Lock()
							recreateOnTimer = true
							mx.Unlock()
							continue
						}
					} else {
						logrus.Infof("resubscribe to OracleRequest on error: %v", err)
						sub = event.Resubscribe(3*time.Second, func(ctx context.Context) (event.Subscription, error) {
							return client.BridgeFilterer.WatchOracleRequest(opt, channel)
						})
					}
				}
			case e := <-channel:
				if e != nil {
					logrus.Infof("BridgeEvent received on chainId: %s", e.Chainid.String())
					currentTopic := common2.ToHex(e.Raw.TxHash)
					logrus.Debugf("currentTopic %s", currentTopic)
					if sendTopic, err := n.P2PPubSub.JoinTopic(currentTopic); err != nil {
						logrus.WithFields(logrus.Fields{
							field.CainId:              e.Chainid,
							field.ConsensusRendezvous: currentTopic,
						}).Error(fmt.Errorf("join topic error: %w", err))
						continue reqLoop
					} else {
						defer sendTopic.Close()
						go func(topic *pubSub.Topic) {
							defer func() {
								if err := topic.Close(); err != nil {
									logrus.WithFields(logrus.Fields{
										field.CainId:              e.Chainid,
										field.ConsensusRendezvous: currentTopic,
									}).Error(fmt.Errorf("close topic error: %w", err))
								}
								logrus.Tracef("chainId %s topic %s closed", e.Chainid.String(), topic.String())
							}()

							ctx, cancel := context.WithDeadline(n.Ctx, time.Now().Add(20*time.Second))
							defer cancel()
							for {

								p2pSub, err := topic.Subscribe()
								if err != nil {
									logrus.WithFields(logrus.Fields{
										field.CainId:              e.Chainid,
										field.ConsensusRendezvous: currentTopic,
									}).Error(fmt.Errorf("subscribe error: %w", err))
									return
								}
								defer p2pSub.Cancel()
								n.P2PPubSub.InitializePubSubWithTopic(n.Host, topic.String())
								parts, enaugh := n.EnaughTopicParticipants(topic)

								if enaugh {
									publicKeys, err := GetPubKeysFromContract(client)
									if err != nil {
										return
									}
									n.session = n.NewBLSSession(topic, publicKeys, parts)
									n.session.Leader, err = libp2p.RelayerLeaderNode(n.session.Topic.String(), parts)
									if err != nil {
										logrus.Errorf("RelayerLeaderNode error: %v", err)
										return
									}

									logrus.Warnf("SESSION LEADER %v", n.session.Leader)
									//logrus.Warnf("SESSION Participants %v", parts)

									wg := new(sync.WaitGroup)
									wg.Add(1)
									go n.StartProtocolByOracleRequest(e, wg)
									wg.Wait()
								}
								if ctx.Err() != nil {
									logrus.Warnf("Not enaugh participants %d , %v", len(parts), ctx.Err())
									break
								}
								time.Sleep(300 * time.Millisecond)
							}
						}(sendTopic)
					}
				}
			case <-n.Ctx.Done():
				err = ErrContextDone
				break reqLoop
			}
		}
	}(&sub, &client)
	return
}

func (n *Node) EnaughTopicParticipants(topic *pubSub.Topic) ([]peer.ID, bool) {
	topicParticipants := topic.ListPeers()
	topicParticipants = append(topicParticipants, n.Host.ID())
	logrus.Tracef("len(topicParticipants) = [ %d ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.Dht.RoutingTable().ListPeers()))
	if len(topicParticipants) > minConsensusNodesCount && len(topicParticipants) > len(n.P2PPubSub.ListPeersByTopic(config.Bridge.Rendezvous))/2+1 {
		return topicParticipants, true
	}
	return topicParticipants, false
}

func (n *Node) ReceiveRequestV2(event *wrappers.BridgeOracleRequest) (receipt *types.Receipt, err error) {
	logrus.Infof("event.Bridge: %v, event.Chainid: %v, event.OppositeBridge: %v, event.ReceiveSide: %v, event.Selector: %v, event.RequestType: %v",
		event.Bridge, event.Chainid, event.OppositeBridge, event.ReceiveSide, common2.BytesToHex(event.Selector), event.RequestType)

	client, _, err := n.GetNodeClientOrRecreate(event.Chainid)
	if err != nil {
		logrus.WithFields(
			field.ListFromBridgeOracleRequest(event),
		).Error(fmt.Errorf("can not get client on error: %w", err))
		return
	}

	logrus.Infof("going to make this call in %s chain", client.ChainCfg.ChainId.String())
	/** Invoke bridge on another side */
	instance, err := wrappers.NewBridge(event.OppositeBridge, client.EthClient)
	if err != nil {
		logrus.WithFields(
			field.ListFromBridgeOracleRequest(event),
		).Error(fmt.Errorf("invoke opposite bridge error: %w", err))
	}
	n.nonceMx.Lock()

	var txHash *common.Hash

	if client.ChainCfg.UseGsn && n.gsnClient != nil {
		hash, err := forward.BridgeRequestV2(n.gsnClient, event.Chainid, n.signerKey, client.ChainCfg.BridgeAddress, event.RequestId, event.Selector, event.ReceiveSide, event.Bridge)
		if err != nil {
			err = fmt.Errorf("ReceiveRequestV2 gsn error:%w", err)
			logrus.WithFields(
				field.ListFromBridgeOracleRequest(event),
			).Error(err)
			n.nonceMx.Unlock()
			return nil, err
		} else {
			txHash = &hash
		}
	} else {
		txOpts := common2.CustomAuth(client.EthClient, n.signerKey)
		/** Invoke bridge on another side */
		tx, err := instance.ReceiveRequestV2(txOpts, event.RequestId, event.Selector, event.ReceiveSide, event.Bridge)
		if err != nil {
			err = fmt.Errorf("ReceiveRequestV2 error:%w", err)
			logrus.WithFields(
				field.ListFromBridgeOracleRequest(event),
			).Error(err)
			n.nonceMx.Unlock()
			return nil, err
		} else {
			hash := tx.Hash()
			txHash = &hash
		}
	}
	n.nonceMx.Unlock()
	if txHash != nil {
		receipt, err = helpers.WaitTransactionDeadline(client.EthClient, *txHash, 30*time.Second)
		if err != nil || receipt == nil {
			err = fmt.Errorf("ReceiveRequestV2 Failed on error: %w", err)
			logrus.WithFields(logrus.Fields{
				field.BridgeRequest: field.ListFromBridgeOracleRequest(event),
				field.TxId:          txHash.Hex(),
			}).Error()
			return nil, err
		}
	}
	return
}

func (n Node) InitializeCommonPubSub() (p2pPubSub *bls_consensus.Protocol) {
	return new(bls_consensus.Protocol)
}

// DiscoverByRendezvous	Announce your presence in network using a rendezvous point
// With the DHT set up, it’s time to discover other peers
// The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
// It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.
// routingDiscovery.Advertise makes this node announce that it can provide a value for the given key.
// Where a key in this case is rendezvousString. Other peers will hit the same key to find other peers.
func (n Node) DiscoverByRendezvous(wg *sync.WaitGroup, rendezvous string) {

	defer wg.Done()
	//	The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
	//	It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.

	// TODO: When TTL elapsed should check presence in network
	// QUEST: What's happening with presence in network when node goes down (in DHT table, in while other nodes is trying to connect)
	logrus.Printf("Announcing ourselves with rendezvous [%s] ...", rendezvous)
	var routingDiscovery = discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, rendezvous)
	logrus.Printf("Successfully announced! n.Host.ID():%s ", n.Host.ID().Pretty())

	ticker := time.NewTicker(config.Bridge.TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logrus.Tracef("Looking advertised peers by rendezvous %s....", rendezvous)
			// TODO: enhance, because synchronous
			peers, err := discovery.FindPeers(n.Ctx, routingDiscovery, rendezvous)
			if err != nil {
				logrus.Fatal("FindPeers ", err)
			}
			ctn := 0
			for _, p := range peers {
				if p.ID == n.Host.ID() {
					continue
				}
				logrus.Tracef("Discovery: FoundedPee %v, isConnected: %t", p, n.Host.Network().Connectedness(p.ID) == network.Connected)
				// TODO: add into if: "&&  n.n.DiscoveryPeers.contains(p.ID)"

				if n.Host.Network().Connectedness(p.ID) != network.Connected {
					_, err = n.Host.Network().DialPeer(n.Ctx, p.ID)
					if err != nil {
						logrus.WithField(field.PeerId, p.ID.Pretty()).
							Error(errors.New("connect to peer was unsuccessful on discovery"))

					} else {
						logrus.Tracef("Discovery: connected to peer %s", p.ID.Pretty())
					}
				} else {
					// store peer uptime
					n.uptimeRegistry.Get(p.ID.Pretty()).Mark(uint64(config.Bridge.TickerInterval.Milliseconds() / 1000))
				}

				if n.Host.Network().Connectedness(p.ID) == network.Connected {
					ctn++
				}

			}
			logrus.Tracef("Total amout discovered and connected peers: %d", ctn)

		case <-n.Ctx.Done():
			return
		}
	}
}

/*func (n *Node) UptimeSchedule(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
	uptimeLoop:
		for t := range schedule.TimeStream(n.Ctx, time.Time{}, config.Bridge.UptimeReportInterval) {
			currentTopic := schedule.TimeToTopicName("uptime", t)
			if uptimeTopic, err := n.P2PPubSub.JoinTopic(currentTopic); err != nil {
				logrus.WithFields(logrus.Fields{
					field.ConsensusRendezvous: currentTopic,
				}).Error(fmt.Errorf("join uptime topic error: %w", err))
				continue uptimeLoop
			} else {
				go func(topic *pubSub.Topic) {
					defer func() {
						if err := topic.Close(); err != nil {
							logrus.WithFields(logrus.Fields{
								field.ConsensusRendezvous: currentTopic,
							}).Error(fmt.Errorf("close uptime topic error: %w", err))
						}
						logrus.Tracef("uptime topic %s closed", topic.String())
					}()
					p2pSub, err := topic.Subscribe()
					if err != nil {
						logrus.WithFields(logrus.Fields{
							field.ConsensusRendezvous: currentTopic,
						}).Error(fmt.Errorf("uptime subscribe error: %w", err))
						return
					}
					defer p2pSub.Cancel()
					logrus.Infoln("uptime.Topic: ", p2pSub.Topic())
					logrus.Infoln("uptime.ListPeers(): ", topic.ListPeers())

					var nodeBls *model.Node
					nodeBls, err = n.NewBLSNode(uptimeTopic, n.Clients[config.Bridge.Chains[0].ChainId.String()])
					if err != nil {
						err = fmt.Errorf("uptime create new bls node error: %w ", err)
						logrus.WithFields(logrus.Fields{
							field.ConsensusRendezvous: currentTopic,
						}).Error(err)
						return
					}
					if nodeBls != nil {
						wg := new(sync.WaitGroup)
						wg.Add(1)
						go n.startUptimeProtocol(t, wg, nodeBls)
						wg.Wait()
					}
				}(uptimeTopic)
			}
		}

	}()
}*/

/*func (n *Node) startUptimeProtocol(t time.Time, wg *sync.WaitGroup, nodeBls *model.Node) {
	defer wg.Done()
	// TODO AdvanceWithTopic switched off for now - to restore after epoch and setup phase intagrated
	//consensusChannel := make(chan bool)
	logrus.Tracef("uptimeRendezvous %v LEADER %v", nodeBls.Topic, nodeBls.Leader)
	wg.Add(1)
	go nodeBls.AdvanceStep(0)
	time.Sleep(time.Second)
	//wg.Add(1)
	//go nodeBls.WaitForProtocolMsg(consensusChannel, wg)
	//consensus := <-consensusChannel
	consensus := true
	if consensus == true {
		//logrus.Tracef("Starting uptime Leader election !!!")
		//leaderPeerId, err := libp2p.RelayerLeaderNode(nodeBls.Topic.String(), nodeBls.Participants)
		//if err != nil {
		//	panic(err)
		//}
		logrus.Debugf("LEADER id %s my ID %s", nodeBls.Leader.Pretty(), n.Host.ID().Pretty())
		if n.session.Leader.Pretty() == n.Host.ID().Pretty() {
			logrus.Infof("LEADER IS %s", n.session.Leader.Pretty())
			time.Sleep(3 * time.Second) // sleep for rpc uptime servers can start
			logrus.Info("LEADER going to get uptime from over nodes")
			uptimeLeader := uptime.NewLeader(n.Host, n.uptimeRegistry, n)
			uptimeData := uptimeLeader.Uptime()
			logrus.Infof("uptime data: %v", uptimeData)
			logrus.Infof("LEADER going to reset uptime %s", t.String())
			uptimeLeader.Reset()
			logrus.Infof("The END of Protocol")
		} else {
			logrus.Debug("start uptime server")
			if uptimeServer, err := uptime.NewServer(n.Host, leaderPeerId, n.uptimeRegistry); err != nil {
				logrus.WithFields(logrus.Fields{
					field.ConsensusRendezvous: nodeBls.Topic,
				}).Error(fmt.Errorf("can not init uptime server on error: %w", err))
			} else {
				uptimeServer.WaitForReset()
				logrus.Debug("uptime server stopped")

			}
			logrus.Debug("The END of Protocol")
		}
	}

}*/

func (n *Node) GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error) {
	if c, _, err := n.GetNodeClientOrRecreate(chainId); err != nil {

		return nil, err
	} else {

		return &c.Forwarder, err
	}
}

func (n *Node) GetForwarderAddress(chainId *big.Int) (common.Address, error) {
	if c, ok := n.Clients[chainId.String()]; !ok {

		return common.Address{}, fmt.Errorf("invalid chain [%s]", chainId.String())
	} else {

		return c.ChainCfg.ForwarderAddress, nil
	}
}

const ChanLen = 500

const (
	BlsSetupPhase = iota
	BlsSetupParts
)

type MessageBlsSetup struct {
	Source             int // NodeID of message's source
	MsgType            int // Type of message
	MembershipKeyParts []bls.Signature
}

// BlsSetup handles BLS setup phase and returns membership key as a result
func (n *Node) BlsSetup(wg *sync.WaitGroup) bls.Signature {
	ctx, cancel := context.WithCancel(n.Ctx)
	defer cancel()
	mutex := &sync.Mutex{}
	np := len(n.PublicKeys)
	anticoefs := bls.CalculateAntiRogueCoefficients(n.PublicKeys)
	receivedMembershipKeyMask := *big.NewInt((1 << np) - 1)
	membershipKey := bls.ZeroSignature()
	msgChan := make(chan *[]byte, ChanLen)

	for ctx.Err() == nil {
		rcvdMsg := n.P2PPubSub.Receive(ctx)
		if rcvdMsg == nil {
			continue
		}
		msgChan <- rcvdMsg
		wg.Add(1)
		go func() {
			defer wg.Done()
			msgBytes := <-msgChan
			var msg MessageBlsSetup
			if err := json.Unmarshal(*msgBytes, &msg); err != nil {
				logrus.Errorf("Unable to decode received message, skipped: %v %d", *msgBytes, n.Id)
				return
			}

			switch msg.MsgType {
			case BlsSetupPhase:
				membershipKeyParts := make([]bls.Signature, len(n.PublicKeys))
				for i := range membershipKeyParts {
					membershipKeyParts[i] = n.PrivKey.GenerateMembershipKeyPart(byte(i), n.EpochPublicKey, anticoefs[n.Id])
				}
				outmsg := MessageBlsSetup{
					Source:             n.Id,
					MsgType:            BlsSetupParts,
					MembershipKeyParts: membershipKeyParts,
				}
				msgBytes, _ := json.Marshal(outmsg)
				n.P2PPubSub.Broadcast(msgBytes)

			case BlsSetupParts:
				logrus.Infof("Membership Key parts of node %d received by node %d", msg.Source, n.Id)
				part := msg.MembershipKeyParts[n.Id]
				if !part.VerifyMembershipKeyPart(n.EpochPublicKey, n.PublicKeys[msg.Source], anticoefs[msg.Source], byte(n.Id)) {
					logrus.Errorf("Failed to verify membership key from node %d on node %d", msg.Source, n.Id)
				}
				mutex.Lock()
				membershipKey = membershipKey.Aggregate(part)
				receivedMembershipKeyMask.SetBit(&receivedMembershipKeyMask, msg.Source, 0)
				if receivedMembershipKeyMask.Sign() == 0 {
					cancel()
				}
				mutex.Unlock()
			}
		}()
	}
	return membershipKey
}

func (n *Node) StartEpoch(client Client, nodeIdAddress common.Address, rendezvous string) error {
	nodeFromContract, err := client.NodeRegistry.GetNode(nodeIdAddress)
	if err != nil {
		return err
	}
	logrus.Infof("Node address: %x nodeAddress from contract: %x", nodeIdAddress, nodeFromContract.NodeIdAddress)
	if nodeFromContract.NodeIdAddress != nodeIdAddress {
		logrus.Fatalf("Peer addresses mismatch. Contract: %s Local file: %s", nodeFromContract.NodeIdAddress, nodeIdAddress)
	}

	publicKeys, err := GetPubKeysFromContract(client)
	if err != nil {
		return err
	}
	anticoefs := bls.CalculateAntiRogueCoefficients(publicKeys)
	aggregatedPublicKey := bls.AggregatePublicKeys(publicKeys, anticoefs)
	aggregatedPublicKey.Marshal() // to avoid data race because Marshal() wants to normalize the key for the first time

	n.Id = int(nodeFromContract.NodeId.Int64())
	n.PublicKeys = publicKeys
	n.EpochPublicKey = aggregatedPublicKey

	// for n.P2PPubSub.TopicObj() == nil || len(n.P2PPubSub.TopicObj().ListPeers()) < len(publicKeys)-1 {
	// 	if n.P2PPubSub.TopicObj() == nil {
	// 		logrus.Info("Waiting to join the initial topic")
	// 	} else {
	// 		logrus.Info("Waiting for all members to come, sendTopic.ListPeers(): ", n.P2PPubSub.TopicObj().ListPeers())
	// 	}
	// 	time.Sleep(300 * time.Millisecond)
	// }

	if n.Id == 0 {
		logrus.Info("Preparing to start BLS setup phase...")
		go func() {
			// TODO: wait until every participant is online
			time.Sleep(19000 * time.Millisecond)

			// Fire the setip phase
			msg := MessageBlsSetup{MsgType: BlsSetupPhase}
			msgBytes, _ := json.Marshal(msg)
			n.P2PPubSub.Broadcast(msgBytes)
		}()
	}
	wg := &sync.WaitGroup{}
	n.MembershipKey = n.BlsSetup(wg)
	return nil
}
