package bridge

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/libp2p/go-flow-metrics"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/uptime"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/schedule"
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
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/helpers"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/pub_sub_bls/libp2p_pubsub"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/pub_sub_bls/modelBLS"
	messageSigPb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/pub_sub_bls/protobuf/messageWithSig"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/util/encoding"
)

var ErrContextDone = errors.New("interrupt on context done")

const minConsensusNodesCount = 5

type Node struct {
	base.Node
	cMx            *sync.Mutex
	Clients        map[string]Client
	nonceMx        *sync.Mutex
	P2PPubSub      *libp2p_pubsub.Libp2pPubSub
	PrivKey        kyber.Scalar
	uptimeRegistry *flow.MeterRegistry
	gsnClient      *gsn.Client
}

func (n Node) StartProtocolByOracleRequest(event *wrappers.BridgeOracleRequest, wg *sync.WaitGroup, nodeBls *modelBLS.Node) {
	defer wg.Done()
	consensusChannel := make(chan bool)
	logrus.Tracef("сurrentRendezvous %v LEADER %v", nodeBls.CurrentRendezvous, nodeBls.Leader)
	wg.Add(1)
	go nodeBls.AdvanceWithTopic(0, nodeBls.CurrentRendezvous, wg)
	wg.Add(1)
	go nodeBls.WaitForMsgNEW(consensusChannel, wg)
	consensus := <-consensusChannel
	if consensus == true {
		logrus.Tracef("Starting Leader election !!!")
		leaderPeerId, err := libp2p.RelayerLeaderNode(nodeBls.CurrentRendezvous, nodeBls.Participants)
		if err != nil {
			panic(err)
		}
		logrus.Infof("LEADER IS %v", leaderPeerId)
		logrus.Debugf("LEADER id %s my ID %s", nodeBls.Leader.Pretty(), n.Host.ID().Pretty())
		if leaderPeerId.Pretty() == n.Host.ID().Pretty() {
			logrus.Info("LEADER going to Call external chain contract method")
			_, err := n.ReceiveRequestV2(event)
			if err != nil {
				logrus.Errorf("%v", err)
			}
		}
	}
	logrus.Println("The END of Protocol")
}

func (n Node) nodeExists(client Client, nodeIdAddress common.Address) bool {
	node, err := common2.GetNode(client.EthClient, client.ChainCfg.NodeListAddress, nodeIdAddress)
	if err != nil || node.NodeWallet == common.HexToAddress("0") {
		return false
	}
	return true

}

func (n Node) GetPubKeysFromContract(client Client) (publicKeys []kyber.Point, err error) {

	if client.EthClient == nil {
		return nil, ErrGetEthClient
	}

	suite := pairing.NewSuiteBn256()
	publicKeys = make([]kyber.Point, 0)
	nodes, err := common2.GetNodesFromContract(client.EthClient, client.ChainCfg.NodeListAddress)
	if err != nil {
		return
	}
	for _, node := range nodes {
		p, err := encoding.ReadHexPoint(suite, strings.NewReader(node.BlsPubKey))
		if err != nil {
			panic(err)
		}
		publicKeys = append(publicKeys, p)
	}
	return
}

func (n Node) KeysFromFilesByConfigName(name string) (prvKey kyber.Scalar, err error) {

	nodeKeyFile := "keys/" + name + "-bn256.key"
	prvKey, err = common2.ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return
	}

	return
}

func (n Node) NewBLSNode(topic *pubSub.Topic, client Client) (blsNode *modelBLS.Node, err error) {

	publicKeys, err := n.GetPubKeysFromContract(client)
	if err != nil {
		return
	}

	suite := pairing.NewSuiteBn256()

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	if !n.nodeExists(client, nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())

	} else {
		logrus.Tracef("Host.ID() %v ", n.Host.ID())
		node, err := client.NodeList.GetNode(nodeIdAddress)
		if err != nil {
			return nil, err
		}

		mask, err := sign.NewMask(suite, publicKeys, nil)
		if err != nil {
			return nil, err
		}
		blsNode = func() *modelBLS.Node {
			ctx, cancel := context.WithTimeout(n.Ctx, 10*time.Second)
			defer cancel()
			for {
				topicParticipants := topic.ListPeers()
				topicParticipants = append(topicParticipants, n.Host.ID())
				// logrus.Tracef("len(topicParticipants) = [ %d ] len(n.DiscoveryPeers)/2+1 = [ %v ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants) /*, len(n.DiscoveryPeers)/2+1*/, len(n.Dht.RoutingTable().ListPeers()))
				logrus.Tracef("len(topicParticipants) = [ %d ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.Dht.RoutingTable().ListPeers()))
				if len(topicParticipants) > minConsensusNodesCount && len(topicParticipants) > len(n.P2PPubSub.ListPeersByTopic(config.Bridge.Rendezvous))/2+1 {
					blsNode = &modelBLS.Node{
						Id:                int(node.NodeId),
						TimeStep:          0,
						ThresholdWit:      len(topicParticipants)/2 + 1,
						ThresholdAck:      len(topicParticipants)/2 + 1,
						Acks:              0,
						ConvertMsg:        &messageSigPb.Convert{},
						Comm:              n.P2PPubSub,
						History:           make([]modelBLS.MessageWithSig, 0),
						Signatures:        make([][]byte, len(publicKeys)),
						SigMask:           mask,
						PublicKeys:        publicKeys,
						PrivateKey:        n.PrivKey,
						Suite:             suite,
						Participants:      topicParticipants,
						CurrentRendezvous: topic.String(),
						Leader:            "",
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
}

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
				logrus.Infof("ReceiveRequest: %v %v %v %v", e.ReqId, e.ReceiveSide, e.BridgeFrom, e.SenderSide)
				// TODO disconnect from topic
				/** TODO:
				Is transaction true, otherwise repeat to invoke tx, err := instance.ReceiveRequestV2(auth)
				*/

			}
		}
	}()
	return

}

func (n *Node) ListenNodeOracleRequest(channel chan *wrappers.BridgeOracleRequest, wg *sync.WaitGroup, chainId *big.Int) (err error) {
	opt := &bind.WatchOpts{}
	recreateOnTimer := false
	checkClientTimer := time.NewTicker(10 * time.Second)
	mx := new(sync.Mutex)
	var sub event.Subscription
	client, clientRecreated, err := n.GetNodeClientOrRecreate(chainId)
	if err != nil {
		if errors.Is(err, ErrGetEthClient) {
			recreateOnTimer = true
			sub = event.NewSubscription(func(i <-chan struct{}) error {
				return nil
			})
		} else {
			return err
		}
	} else {
		sub, err = client.BridgeFilterer.WatchOracleRequest(opt, channel)
		if err != nil {
			logrus.Errorf("WatchOracleRequest can't %v", err)
			return
		}
	}

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
						clientRecreated = false
						*clientPtr, clientRecreated, err = n.GetNodeClientOrRecreate(chainId)
						if errors.Is(err, ErrGetEthClient) {
							recreateOnTimer = true
							return err
						} else if err != nil {
							logrus.WithField(field.CainId, chainId.String()).
								Error(fmt.Errorf("can not get client for network [%s] on error:%w",
									chainId.String(), err))
							recreateOnTimer = true
							time.Sleep(1 * time.Second)
							return err
						} else {
							logrus.Infof("client for network [%s] restored on timer. recreated: %v",
								chainId.String(), clientRecreated)

							if clientRecreated {
								logrus.Info("resubscribe to OracleRequest recreated client on timer")
								*subPtr, err = clientPtr.BridgeFilterer.WatchOracleRequest(opt, channel)
								if err != nil {
									logrus.Error(fmt.Errorf("WatchOracleRequest can't %w", err))
									recreateOnTimer = true
									time.Sleep(1 * time.Second)
									return err
								} else {
									recreateOnTimer = false
									return nil
								}
							} else {
								logrus.Info("resubscribe to OracleRequest")
								*subPtr = event.Resubscribe(3*time.Second, func(ctx context.Context) (event.Subscription, error) {
									return clientPtr.BridgeFilterer.WatchOracleRequest(opt, channel)
								})
								recreateOnTimer = false
								return nil
							}

						}

					}
					return nil
				}(); err != nil {
					continue
				}
			case err := <-(*subPtr).Err():
				if err != nil {
					logrus.Error(fmt.Errorf("OracleRequest subscription error: %w chainId: %s", err, chainId.String()))
					func() {
						mx.Lock()
						defer mx.Unlock()
						clientRecreated := false
						*clientPtr, clientRecreated, err = n.GetNodeClientOrRecreate(chainId)
						if err != nil {
							logrus.WithField(field.CainId, chainId.String()).
								Error(fmt.Errorf("can not get client for network [%s] on error:%w",
									chainId.String(), err))
							time.Sleep(1 * time.Second)
							recreateOnTimer = true
							return
						} else {
							logrus.Infof("client for network [%s] restored after sub err: %v, recreated: %v, client: %v",
								chainId.String(), err, clientRecreated, *clientPtr)

							if clientRecreated {
								logrus.Infof("subscribe to OracleRequest on recreated client on sub err: %v", err)
								*subPtr, err = client.BridgeFilterer.WatchOracleRequest(opt, channel)
								if err != nil {
									logrus.Error(fmt.Errorf("WatchOracleRequest can't %w", err))
									time.Sleep(1 * time.Second)
									recreateOnTimer = true
									return
								} else {
									recreateOnTimer = false

									return
								}
							} else {
								logrus.Infof("resubscribe to OracleRequest on error: %v", err)
								*subPtr = event.Resubscribe(3*time.Second, func(ctx context.Context) (event.Subscription, error) {
									return clientPtr.BridgeFilterer.WatchOracleRequest(opt, channel)
								})
								recreateOnTimer = false
							}
						}
					}()
				}
			case e := <-channel:
				if e != nil {
					logrus.Debugf("receive event tx: %s", e.Raw.TxHash.Hex())
					if !n.IsClientReady(e.Chainid) {
						logrus.WithFields(logrus.Fields{
							field.BridgeRequest: field.ListFromBridgeOracleRequest(e),
							field.TxId:          e.Raw.TxHash.Hex(),
						}).Errorf("CANCEL PROTOCOL on chain [%s] because network client not ready", e.Chainid)
						mx.Lock()
						recreateOnTimer = true
						mx.Unlock()
						continue
					}
					logrus.Infof("going to InitializePubSubWithTopicAndPeers on chainId: %s", e.Chainid.String())
					currentTopic := common2.ToHex(e.Raw.TxHash)
					logrus.Debugf("currentTopic %s", currentTopic)
					if sendTopic, err := n.P2PPubSub.JoinTopic(currentTopic); err != nil {
						logrus.WithFields(logrus.Fields{
							field.CainId:              e.Chainid,
							field.ConsensusRendezvous: currentTopic,
						}).Error(fmt.Errorf("join topic error: %w", err))
						continue reqLoop
					} else {
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
							p2pSub, err := topic.Subscribe()
							if err != nil {
								logrus.WithFields(logrus.Fields{
									field.CainId:              e.Chainid,
									field.ConsensusRendezvous: currentTopic,
								}).Error(fmt.Errorf("subscribe error: %w", err))
								return
							}
							defer p2pSub.Cancel()
							logrus.Println("p2pSub.Topic: ", p2pSub.Topic())
							logrus.Println("sendTopic.ListPeers(): ", sendTopic.ListPeers())

							var nodeBls *modelBLS.Node
							nodeBls, err = n.NewBLSNode(sendTopic, *clientPtr)
							if err != nil {
								logrus.WithFields(logrus.Fields{
									field.CainId:              e.Chainid,
									field.ConsensusRendezvous: currentTopic,
								}).Error(fmt.Errorf("create bls node error: %w ", err))
								return
							}
							if nodeBls != nil {
								wg := new(sync.WaitGroup)
								wg.Add(1)
								go n.StartProtocolByOracleRequest(e, wg, nodeBls)
								wg.Wait()
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

func (n *Node) ReceiveRequestV2(event *wrappers.BridgeOracleRequest) (receipt *types.Receipt, err error) {
	logrus.Infof("event.Bridge: %v, event.Chainid: %v, event.OppositeBridge: %v, event.ReceiveSide: %v, event.Selector: %v, event.RequestType: %v",
		event.Bridge, event.Chainid, event.OppositeBridge, event.ReceiveSide, common2.BytesToHex(event.Selector), event.RequestType)

	client, err := n.GetNodeClient(event.Chainid)
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
	txOpts := common2.CustomAuth(client.EthClient, client.EcdsaKey)
	/** Invoke bridge on another side */
	tx, err := instance.ReceiveRequestV2(txOpts, event.RequestId, event.Selector, event.ReceiveSide, event.Bridge)
	if err != nil {
		logrus.WithFields(
			field.ListFromBridgeOracleRequest(event),
		).Error(fmt.Errorf("ReceiveRequestV2 error:%w", err))
	}
	n.nonceMx.Unlock()
	if tx != nil {
		receipt, err = helpers.WaitTransaction(client.EthClient, tx)
		if err != nil || receipt == nil {
			err = fmt.Errorf("ReceiveRequestV2 Failed on error: %w", err)
			logrus.WithFields(logrus.Fields{
				field.BridgeRequest: field.ListFromBridgeOracleRequest(event),
				field.TxId:          tx.Hash().Hex(),
			}).Error()
			return nil, err
		}
	}
	return
}

func (n Node) Discover(topic string) {
	go libp2p.Discover(n.Ctx, n.Host, n.Dht, topic, 3*time.Second)
}

func (n Node) InitializeCommonPubSub() (p2pPubSub *libp2p_pubsub.Libp2pPubSub) {
	return new(libp2p_pubsub.Libp2pPubSub)
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
func (n *Node) UptimeSchedule(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		chainId := big.NewInt(int64(config.Bridge.Chains[0].Id))
	uptimeLoop:
		for t := range schedule.TimeStream(n.Ctx, time.Time{}, config.Bridge.UptimeReportInterval) {
			if !n.IsClientReady(chainId) {
				logrus.WithFields(logrus.Fields{}).Errorf("CANCEL uptime PROTOCOL on chain [%s] because network client not ready", chainId)
				continue
			}

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

					var nodeBls *modelBLS.Node
					nodeBls, err = n.NewBLSNode(uptimeTopic, n.Clients[chainId.String()])
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
}

func (n *Node) startUptimeProtocol(t time.Time, wg *sync.WaitGroup, nodeBls *modelBLS.Node) {
	defer wg.Done()
	consensusChannel := make(chan bool)
	logrus.Tracef("uptimeRendezvous %v LEADER %v", nodeBls.CurrentRendezvous, nodeBls.Leader)
	wg.Add(1)
	go nodeBls.AdvanceWithTopic(0, nodeBls.CurrentRendezvous, wg)
	wg.Add(1)
	go nodeBls.WaitForMsgNEW(consensusChannel, wg)
	consensus := <-consensusChannel
	if consensus == true {
		logrus.Tracef("Starting uptime Leader election !!!")
		leaderPeerId, err := libp2p.RelayerLeaderNode(nodeBls.CurrentRendezvous, nodeBls.Participants)
		if err != nil {
			panic(err)
		}
		logrus.Debugf("LEADER id %s my ID %s", nodeBls.Leader.Pretty(), n.Host.ID().Pretty())
		if leaderPeerId.Pretty() == n.Host.ID().Pretty() {
			logrus.Infof("LEADER IS %s", leaderPeerId.Pretty())
			time.Sleep(3 * time.Second) // sleep for rpc uptime servers can start
			logrus.Info("LEADER going to get uptime from over nodes")
			uptimeLeader := uptime.NewLeader(n.Host, n.uptimeRegistry, nodeBls.Participants)
			uptimeData := uptimeLeader.Uptime()
			logrus.Infof("uptime data: %v", uptimeData)
			logrus.Infof("LEADER going to reset uptime %s", t.String())
			uptimeLeader.Reset()
			logrus.Infof("The END of Protocol")
		} else {
			logrus.Debug("start uptime server")
			if uptimeServer, err := uptime.NewServer(n.Host, leaderPeerId, n.uptimeRegistry); err != nil {
				logrus.WithFields(logrus.Fields{
					field.ConsensusRendezvous: nodeBls.CurrentRendezvous,
				}).Error(fmt.Errorf("can not init uptime server on error: %w", err))
			} else {
				uptimeServer.WaitForReset()
				logrus.Debug("uptime server stopped")

			}
			logrus.Debug("The END of Protocol")
		}
	}

}

func (n *Node) GetDht() *dht.IpfsDHT {
	return n.Dht
}
