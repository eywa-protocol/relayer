package bridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/ledger"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/extChains"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/forward"
	libp2p2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/contracts"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/libp2p/go-flow-metrics"
	bls_consensus "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/model"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/rpc/gsn"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/base"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p-core/network"
	discovery "github.com/libp2p/go-libp2p-discovery"
	pubSub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
	messageSigPb "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/protobuf/message"
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
	contracts.Registry
	contracts.Forwarder
	contracts.Bridge
	rendezvous     string
	clients        *extChains.Clients
	nonceMx        *sync.Mutex
	P2PPubSub      *bls_consensus.Protocol
	signerKey      *ecdsa.PrivateKey
	PrivKey        bls.PrivateKey
	uptimeRegistry *flow.MeterRegistry
	gsnClient      *gsn.Client
	BlsNodeId      int
	Ledger         *ledger.Ledger
	EpochKeys
}

func (n Node) StartProtocolByOracleRequest(event *wrappers.BridgeOracleRequest, session *model.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	consensusChannel := make(chan bool)
	session.AdvanceStep(0)
	time.Sleep(time.Second)
	ctx, cancel := context.WithTimeout(session.Ctx, 30*time.Second)
	defer cancel()
	wg.Add(1)
	go session.WaitForProtocolMsg(consensusChannel, wg)
	select {
	case <-consensusChannel:
		err := n.Ledger.CreateBlockFromEvent(*event)
		if err != nil {
			logrus.Error(fmt.Errorf("CreateBlockFromEvent ERROR: %w", err))
		}
		blockHash := n.Ledger.GetCurrentBlockHash()
		hash := blockHash.ToHexString()
		logrus.Printf("BLOCK HASH: %v Currentheight: %d", hash, n.Ledger.GetCurrentBlockHeight())
		logrus.Infof("LEADER:%s == MYMNODE:%s", session.Leader.Pretty(), n.Host.ID().Pretty())
		if session.Leader.Pretty() == n.Host.ID().Pretty() {
			logrus.Info("LEADER going to Call external chain contract method")
			_, err := n.ReceiveRequestV2(event)
			if err != nil {
				logrus.Error(fmt.Errorf("%w", err))
			}
		}
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logrus.Warn("WaitGroup timed out..")
		}
	}
	logrus.Info("The END of Protocol")
}

func (n Node) GetClients() *extChains.Clients {
	return n.clients
}

func (n Node) getNodeBlsId() (id int, err error) {

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))
	if !n.NodeRegistryNodeExists(nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())
		return -1, err
	} else {
		node, err := n.NodeRegistry().GetNode(&bind.CallOpts{}, nodeIdAddress)
		if err != nil {
			return -1, err
		}
		logrus.Tracef("Host.ID() %v ", n.Host.ID())

		id = int(node.NodeId.Int64())

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

func (n Node) RegChainId() *big.Int {

	return new(big.Int).SetUint64(config.Bridge.RegChainId)
}

func (n Node) NewBLSNode(topic *pubSub.Topic) (blsNode *model.Node, err error) {
	publicKeys, err := n.NodeRegistryPublicKeys()
	if err != nil {
		return
	}

	nodeIdAddress := common.BytesToAddress([]byte(n.Host.ID()))

	if !n.NodeRegistryNodeExists(nodeIdAddress) {
		logrus.Errorf("node %x does not exist", n.Host.ID())

	} else {
		logrus.Tracef("Host.ID() %v ", n.Host.ID())
		node, err := n.NodeRegistry().GetNode(&bind.CallOpts{}, nodeIdAddress)
		if err != nil {
			return nil, err
		}

		blsNode = func() *model.Node {
			ctx, cancel := context.WithTimeout(n.Ctx, 30*time.Second)
			defer cancel()
			for {
				topicParticipants := topic.ListPeers()
				topicParticipants = append(topicParticipants, n.Host.ID())
				// logrus.Tracef("len(topicParticipants) = [ %d ] len(n.DiscoveryPeers)/2+1 = [ %v ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants) /*, len(n.DiscoveryPeers)/2+1*/, len(n.Dht.RoutingTable().ListPeers()))
				logrus.Infof("len(topicParticipants) = [ %d ] len(n.Dht.RoutingTable().ListPeers()) = [ %d ]", len(topicParticipants), len(n.Dht.RoutingTable().ListPeers()))
				if len(topicParticipants) > minConsensusNodesCount && len(topicParticipants) > len(n.P2PPubSub.ListPeersByTopic(config.Bridge.Rendezvous))/2+1 {
					leader, _ := libp2p2.RelayerLeaderNode(topic.String(), topicParticipants)

					blsNode = &model.Node{
						Ctx:            n.Ctx,
						Topic:          topic,
						Id:             int(node.NodeId.Int64()),
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
						Leader:         leader,
					}
					break
				}
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					logrus.Warnf("Not enaugh participants %d , %v", len(topicParticipants), ctx.Err())
					break
				}
				time.Sleep(1000 * time.Millisecond)
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

func (n *Node) ReceiveRequestV2(event *wrappers.BridgeOracleRequest) (receipt *types.Receipt, err error) {
	logrus.Infof("event.Bridge: %v, event.Chainid: %v, event.OppositeBridge: %v, event.ReceiveSide: %v, event.Selector: %v, event.RequestType: %v",
		event.Bridge, event.Chainid, event.OppositeBridge, event.ReceiveSide, common2.BytesToHex(event.Selector), event.RequestType)

	client, err := n.clients.GetEthClient(event.Chainid)
	if err != nil {
		logrus.WithFields(
			field.ListFromBridgeOracleRequest(event),
		).Error(fmt.Errorf("can not get client on error: %w", err))
		return
	}

	logrus.Infof("going to make this call in %s chain", event.Chainid.String())
	/** Invoke bridge on another side */
	instance, err := wrappers.NewBridge(event.OppositeBridge, client)
	if err != nil {
		logrus.WithFields(
			field.ListFromBridgeOracleRequest(event),
		).Error(fmt.Errorf("invoke opposite bridge error: %w", err))
	}
	n.nonceMx.Lock()

	var txHash *common.Hash

	if n.CanUseGsn(event.Chainid) && n.gsnClient != nil {
		hash, err := forward.BridgeRequestV2(n.gsnClient, event.Chainid, n.signerKey, event.OppositeBridge, event.RequestId, event.Selector, event.ReceiveSide, event.Bridge)
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
		txOpts, err := client.CallOpt(n.signerKey)
		if err != nil {
			err = fmt.Errorf("get call opts error:%w", err)
			logrus.WithFields(
				field.ListFromBridgeOracleRequest(event),
			).Error(err)
			n.nonceMx.Unlock()
			return nil, err
		}
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
		receipt, err = client.WaitTransaction(*txHash)
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

func (n *Node) InitializeCommonPubSub(rendezvous string) error {
	if pubsub, err := bls_consensus.NewPubSubWithTopic(n.Host, n.GetDht(), rendezvous); err != nil {
		return err
	} else {
		n.rendezvous = rendezvous
		n.P2PPubSub = pubsub

		return nil
	}

}

// DiscoverByRendezvous	Announce your presence in network using a rendezvous point
// With the DHT set up, it’s time to discover other peers
// The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
// It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.
// routingDiscovery.Advertise makes this node announce that it can provide a value for the given key.
// Where a key in this case is rendezvousString. Other peers will hit the same key to find other peers.
func (n *Node) DiscoverByRendezvous(wg *sync.WaitGroup, rendezvous string) {

	//	The Advertise function starts a go-routine that keeps on advertising until the context gets cancelled.
	//	It announces its presence every 3 hours. This can be shortened by providing a TTL (time to live) option as a fourth parameter.

	// TODO: When TTL elapsed should check presence in network
	// QUEST: What's happening with presence in network when node goes down (in DHT table, in while other nodes is trying to connect)
	logrus.Printf("Announcing ourselves with rendezvous [%s] ...", rendezvous)
	routingDiscovery := discovery.NewRoutingDiscovery(n.Dht)
	discovery.Advertise(n.Ctx, routingDiscovery, rendezvous)
	logrus.Printf("Successfully announced! n.Host.ID():%s ", n.Host.ID().Pretty())
	wg.Add(1)
	go func() {
		defer wg.Done()
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
					if p.ID.Pretty() == n.Host.ID().Pretty() {
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
	}()
}

/*func (n *Node) UptimeSchedule(wg *sync.WaitGroup) {
	logrus.Infof("Start uptime report for every %s", config.Bridge.UptimeReportInterval.String())
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

					var nodeBls *modelBLS.Node
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
}

func (n *Node) startUptimeProtocol(t time.Time, wg *sync.WaitGroup, nodeBls *modelBLS.Node) {
	defer wg.Done()
	// TODO AdvanceWithTopic switched off for now - to restore after epoch and setup phase intagrated
	//consensusChannel := make(chan bool)
	logrus.Tracef("uptimeRendezvous %v LEADER %v", nodeBls.CurrentRendezvous, nodeBls.Leader)
	wg.Add(1)
	go nodeBls.AdvanceWithTopic(0, nodeBls.CurrentRendezvous, wg)
	time.Sleep(time.Second)
	//wg.Add(1)
	//go nodeBls.WaitForProtocolMsg(consensusChannel, wg)
	//consensus := <-consensusChannel
	consensus := true
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

			return blsNode
		}()

	}
	return
}
*/

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

	for {
		select {
		case <-ctx.Done():
			break
		default:
			rcvdMsg := n.P2PPubSub.Receive(ctx)
			if rcvdMsg == nil {
				logrus.Info("receive return nil")
				break
			}
			wg.Add(1)
			go func(msgBytes *[]byte) {
				defer wg.Done()
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
			}(rcvdMsg)
		}
	}
	return membershipKey
}

func (n *Node) StartEpoch() error {

	publicKeys, err := n.NodeRegistryPublicKeys()
	if err != nil {
		return err
	}
	anticoefs := bls.CalculateAntiRogueCoefficients(publicKeys)
	aggregatedPublicKey := bls.AggregatePublicKeys(publicKeys, anticoefs)
	aggregatedPublicKey.Marshal() // to avoid data race because Marshal() wants to normalize the key for the first time

	n.Id = n.BlsNodeId
	n.PublicKeys = publicKeys
	n.EpochPublicKey = aggregatedPublicKey

	for n.P2PPubSub.MainTopic() == nil || len(n.P2PPubSub.MainTopic().ListPeers()) < len(publicKeys)-1 { // TODO: configure how many peers to wait for
		if n.P2PPubSub.MainTopic() == nil {
			logrus.Info("Waiting to join the initial topic")
		} else {
			logrus.Info("Waiting for all members to come, topic.ListPeers(): ", n.P2PPubSub.MainTopic().ListPeers())
		}
		time.Sleep(3 * time.Second)
	}

	wg := &sync.WaitGroup{}
	if n.Id == 0 { // TODO: all nodes must be able to send the first message
		logrus.Info("Preparing to start BLS setup phase...")
		// Fire the setip phase
		msg := MessageBlsSetup{MsgType: BlsSetupPhase}
		msgBytes, _ := json.Marshal(msg)
		n.P2PPubSub.Broadcast(msgBytes)
	}

	logrus.Info("Start BLS setup done")
	n.MembershipKey = n.BlsSetup(wg)

	logrus.Info("BLS setup done")
	wg.Wait()
	logrus.Info("start epoch done")
	return nil
}

func (n *Node) HandleOracleRequest(e *wrappers.BridgeOracleRequest, srcChainId *big.Int) {
	logrus.Infof("going to InitializePubSubWithTopicAndPeers on chainId: %s", srcChainId.String())
	topicName := common2.ToHex(e.Raw.TxHash)
	logrus.Debugf("currentTopic %s", topicName)
	if topic, err := n.P2PPubSub.JoinTopic(topicName); err != nil {
		logrus.WithFields(logrus.Fields{
			field.CainId:              srcChainId,
			field.ConsensusRendezvous: topicName,
		}).Error(fmt.Errorf("join topic error: %w", err))
	} else {
		defer func() {
			if err := n.P2PPubSub.LeaveTopic(topicName); err != nil {
				logrus.WithFields(logrus.Fields{
					field.CainId:              e.Chainid,
					field.ConsensusRendezvous: topicName,
				}).Error(fmt.Errorf("close topic error: %w", err))
			}
			logrus.Tracef("chainId %s topic %s closed", e.Chainid.String(), topic.String())
		}()
		err := n.P2PPubSub.Subscribe(topicName)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				field.CainId:              e.Chainid,
				field.ConsensusRendezvous: topicName,
			}).Error(fmt.Errorf("subscribe to topic error: %w", err))
			return
		}
		logrus.Infof("request chain[%s] subscribe to topic: %s", srcChainId.String(), topicName)
		logrus.Infof("request chain[%s] topic peers: %v", srcChainId.String(), topic.ListPeers())

		session, err := n.NewBLSNode(topic)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				field.CainId:              e.Chainid,
				field.ConsensusRendezvous: topicName,
			}).Error(fmt.Errorf("create bls node error: %w ", err))
			return
		}

		if session != nil {
			wg := new(sync.WaitGroup)
			wg.Add(1)
			go n.StartProtocolByOracleRequest(e, session, wg)
			wg.Wait()
		}
	}
}
