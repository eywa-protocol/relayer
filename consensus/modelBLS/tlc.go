package modelBLS

import (
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

const ChanLen = 500

var Logger1 *log.Logger

// Advance will change the step of the node to a new one and then broadcast a message to the network.
func (node *Node) Advance(step int) {
	node.TimeStep = step
	node.Acks = 0
	node.Wits = 0

	// fmt.Printf("node %d , Broadcast in timeStep %d\n", node.Id, node.TimeStep)
	// fmt.Printf("Node ID %d, STEP %d\n", node.Id, node.TimeStep)

	msg := MessageWithSig{
		Source:  node.Id,
		MsgType: Raw,
		Step:    node.TimeStep,
		History: make([]MessageWithSig, 0),
	}

	node.CurrentMsg = msg
	for i := range node.PublicKeys {
		node.Signatures[i].Clear()
	}
	node.SigMask = common.EmptyMask

	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.Broadcast(*msgBytes)
}

// WaitForMsg  waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForMsg(stop int) (err error) {
	mutex := &sync.Mutex{}
	end := false
	msgChan := make(chan *[]byte, ChanLen)
	nodeTimeStep := 0

	isNeedToStop := func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return nodeTimeStep <= stop
	}
	for isNeedToStop() {
		// For now we assume that the underlying receive function is blocking

		mutex.Lock()
		nodeTimeStep = node.TimeStep
		if end {
			mutex.Unlock()
			break
		}
		mutex.Unlock()

		rcvdMsg := node.Comm.Receive()
		if rcvdMsg == nil {
			continue
		}
		msgChan <- rcvdMsg

		go func(nodeTimeStep int) {

			msgBytes := <-msgChan
			msg := node.ConvertMsg.BytesToModelMessage(*msgBytes)

			logrus.Tracef("node %d\n in nodeTimeStep %d;\nReceived MSG with\n msg.Step %d\n MsgType %d source: %d\n", node.Id, nodeTimeStep, msg.Step, msg.MsgType, msg.Source)

			// Used for stopping the execution after some timesteps
			if nodeTimeStep == stop {
				fmt.Println("Break reached by node ", node.Id)
				mutex.Lock()
				end = true
				mutex.Unlock()
				return
			}

			// If the received message is from a lower step, send history to the node to catch up
			if msg.Step < nodeTimeStep {
				if msg.MsgType == Raw {
					msg.MsgType = Catchup
					msg.Step = nodeTimeStep
					msg.History = node.History
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)
				}
				return
			}

			switch msg.MsgType {
			case Wit:

				if msg.Step > nodeTimeStep+1 {
					return
				}

				err := node.verifyThresholdWitnesses(*msg)
				if err != nil {
					logrus.Error(err, " at node ", node.Id, msg.Signature.Marshal())
					return
				}

				if msg.Step == nodeTimeStep+1 { // Node needs to catch up with the message
					// Update nodes local history. Append history from message to local history
					mutex.Lock()
					node.History = append(node.History, *msg)

					// Advance
					node.Advance(msg.Step)
					node.Wits += 1
					mutex.Unlock()
				} else if msg.Step == nodeTimeStep {

					mutex.Lock()
					fmt.Printf("WITS: node %d , %d\n", node.Id, node.Wits)
					// Count message toward the threshold
					node.Wits += 1
					if node.Wits >= node.ThresholdWit {
						// Log the message in history
						node.History = append(node.History, *msg)
						// Advance to next time step
						node.Advance(nodeTimeStep + 1)
					}
					mutex.Unlock()
				}

			case Ack:
				// Checking that the ack is for message of this step
				mutex.Lock()
				if (msg.Step != node.CurrentMsg.Step) || (node.Acks >= node.ThresholdAck) || (node.SigMask.Bit(msg.Source) != 0) {
					mutex.Unlock()
					return
				}
				mutex.Unlock()
				fmt.Printf("node %d received ACK from node %d\n", node.Id, msg.Source)

				err := node.verifyAckSignature(*msg)
				if err != nil {
					logrus.Error(err, " at node ", node.Id, msg.Signature.Marshal())
					return
				}

				// add message's mask to existing mask
				mutex.Lock()
				node.SigMask.Or(&node.SigMask, &msg.Mask)

				// Count acks toward the threshold
				node.Acks += 1
				node.Signatures[msg.Source] = msg.Signature

				if node.Acks >= node.ThresholdAck {
					// Add aggregate signatures to message
					msg.Mask = node.SigMask
					msg.Signature = common.AggregateBlsSignatures(node.Signatures, &node.SigMask)

					// Verify before sending message to others
					if err := node.verifyThresholdWitnesses(*msg); err != nil {
						logrus.Error("verifyThresholdWitnesses ", err, ", node: ", node.Id, ", mask: ", msg.Mask.Text(16))
						//panic(err)
						return
					}

					// Send witnessed message if the acks are more than threshold
					msg.MsgType = Wit

					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)
				}
				mutex.Unlock()

			case Raw:
				if msg.Step > nodeTimeStep+1 {
					return
				} else if msg.Step == nodeTimeStep+1 {
					// Node needs to catch up with the message
					// Update nodes local history. Append history from message to local history
					mutex.Lock()
					node.History = append(node.History, *msg)

					// Advance
					node.Advance(msg.Step)
					mutex.Unlock()
				}

				msg.MsgType = Raw
				msg.Source = -1
				msg.Mask = common.EmptyMask
				msg.Signature.Clear()
				msgBytes := node.ConvertMsg.MessageToBytes(*msg)

				logrus.Debug("Signing message ", *msgBytes, " at node ", node.Id)
				signature := node.PrivateKey.Sign(*msgBytes)

				// Adding signature and ack to message. These fields were empty when message got signed
				msg.MsgType = Ack
				msg.Source = node.Id
				msg.Mask.SetBit(&common.EmptyMask, node.Id, 1)
				msg.Signature = signature
				msgBytes = node.ConvertMsg.MessageToBytes(*msg)
				node.Comm.Send(*msgBytes, msg.Source)

			case Catchup:
				mutex.Lock()
				if msg.Source == node.CurrentMsg.Source && msg.Step > nodeTimeStep {
					fmt.Printf("Catchup: node (%d,step %d), msg(source %d ,step %d)\n", node.Id, node.TimeStep, msg.Source, msg.Step)

					node.History = append(node.History, msg.History[nodeTimeStep:]...)
					node.Advance(msg.Step)

				}
				mutex.Unlock()
			}
		}(nodeTimeStep)

	}
	return err
}

func Popcount(z *big.Int) int {
	var count int
	for _, x := range z.Bits() {
		for x != 0 {
			x &= x - 1
			count++
		}
	}
	return count
}

func (node *Node) verifyThresholdWitnesses(msg MessageWithSig) (err error) {
	// Verify that it's really witnessed by majority of nodes by checking the signature and number of them
	sig := msg.Signature
	mask := msg.Mask

	msg.MsgType = Raw
	msg.Source = -1
	msg.Mask = common.EmptyMask
	msg.Signature.Clear()
	msgBytes := node.ConvertMsg.MessageToBytes(msg)

	if pop := Popcount(&mask); pop < node.ThresholdAck {
		err = fmt.Errorf("Not enough sigantures: %d < %d", pop, node.ThresholdAck)
		return
	}

	aggPubKey := common.AggregateBlsPublicKeys(node.PublicKeys, &mask)

	// Verify message signature
	if sig.Verify(aggPubKey, *msgBytes) == false {
		return fmt.Errorf("Threshold signature mismatch for %v", *msgBytes)
	}
	logrus.Tracef("Aggregated Signature VERIFIED ! ! !")
	return
}

func (node *Node) verifyAckSignature(msg MessageWithSig) (err error) {
	sig := msg.Signature
	source := msg.Source

	msg.MsgType = Raw
	msg.Source = -1
	msg.Mask = common.EmptyMask
	msg.Signature.Clear()
	msgBytes := node.ConvertMsg.MessageToBytes(msg)

	if sig.Verify(node.PublicKeys[source], *msgBytes) == false {
		err = fmt.Errorf("ACK signature mismatch for %v", *msgBytes)
		return
	}
	return
}

// AdvanceWithTopic  will change the step of the node to a new one and then broadcast a message to the network.
func (node *Node) AdvanceWithTopic(step int, topic string, wg *sync.WaitGroup) {
	wg.Done()
	node.TimeStep = step
	node.Acks = 0
	node.Wits = 0

	// fmt.Printf("Node ID %d, STEP %d\n", node.Id, node.TimeStep)

	msg := MessageWithSig{
		Source:  node.Id,
		MsgType: Raw,
		Step:    node.TimeStep,
		History: make([]MessageWithSig, 0),
	}

	node.CurrentMsg = msg
	for i := range node.PublicKeys {
		node.Signatures[i].Clear()
	}
	node.SigMask = common.EmptyMask

	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.Broadcast(*msgBytes)
}

func (node *Node) DisconnectPubSub() {
	node.Comm.Disconnect()

}

// WaitForProtocolMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForProtocolMsg(consensusAgreed chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	mutex := &sync.Mutex{}
	end := false
	msgChan := make(chan *[]byte, ChanLen)
	nodeTimeStep := 0
	stop := 1
	isNeedToStop := func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return nodeTimeStep <= stop
	}
	for isNeedToStop() {
		// For now we assume that the underlying receive function is blocking

		mutex.Lock()
		nodeTimeStep = node.TimeStep
		if end {
			mutex.Unlock()
			break
		}
		mutex.Unlock()

		rcvdMsg := node.Comm.Receive()
		logrus.Trace("rcvdMsg:", rcvdMsg)
		if rcvdMsg == nil {
			node.DisconnectPubSub()
			break
		}
		msgChan <- rcvdMsg
		wg.Add(1)
		go func(nodeTimeStep int) {
			defer wg.Done()
			msgBytes := <-msgChan
			msg := node.ConvertMsg.BytesToModelMessage(*msgBytes)

			if nodeTimeStep == stop {
				logrus.Infof(" Consensus achieved by node %v", node.Id)
				mutex.Lock()
				end = true
				node.TimeStep++
				node.Advance(node.TimeStep)
				mutex.Unlock()
				consensusAgreed <- true
				return
			}

			// If the received message is from a lower step, send history to the node to catch up
			if msg.Step < nodeTimeStep {
				if msg.MsgType == Raw {
					msg.MsgType = Catchup
					msg.Step = nodeTimeStep
					msg.History = node.History
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)

				}
				return
			}

			switch msg.MsgType {
			case Wit:

				if msg.Step > nodeTimeStep+1 {
					return
				}

				err := node.verifyThresholdWitnesses(*msg)
				if err != nil {
					logrus.Error(err, " at node ", node.Id, msg.Signature.Marshal())
					return
				}
				logrus.Debugf("Verified Vit Signature at node %d from node %d", node.Id, msg.Source)

				mutex.Lock()
				node.Wits += 1
				node.TimeStep += 1
				node.Advance(nodeTimeStep + 1)
				mutex.Unlock()

			case Ack:
				// Checking that the ack is for message of this step
				mutex.Lock()
				if (msg.Step != node.CurrentMsg.Step) || (node.Acks >= node.ThresholdAck) || (node.SigMask.Bit(msg.Source) != 0) {
					mutex.Unlock()
					return
				}
				mutex.Unlock()

				err := node.verifyAckSignature(*msg)
				if err != nil {
					logrus.Error(err, " at node ", node.Id, msg.Signature.Marshal())
					return
				}
				logrus.Debugf("Verified Ack Signature at node %d from node %d", node.Id, msg.Source)

				mutex.Lock()
				node.SigMask.Or(&node.SigMask, &msg.Mask)
				logrus.Debugf("Node SigMask Merged: ", node.SigMask.Text(16))

				// Count acks toward the threshold
				node.Acks += 1
				node.Signatures[msg.Source] = msg.Signature

				if node.Acks >= node.ThresholdAck {
					// Add aggregate signatures to message
					msg.Mask = node.SigMask
					msg.Signature = common.AggregateBlsSignatures(node.Signatures, &node.SigMask)

					// Verify before sending message to others
					if err := node.verifyThresholdWitnesses(*msg); err != nil {
						logrus.Error("verifyThresholdWitnesses ", err, ", node: ", node.Id, ", mask: ", msg.Mask.Text(16))
						return
					}

					// Send witnessed message if the acks are more than threshold
					msg.MsgType = Wit

					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)
				}

				mutex.Unlock()

			case Raw:
				if msg.Step > nodeTimeStep+1 {
					return
				} else if msg.Step == nodeTimeStep+1 { // Node needs to catch up with the message
					// Update nodes local history. Append history from message to local history
					mutex.Lock()
					node.History = append(node.History, *msg)

					// Advance
					node.Advance(msg.Step)
					mutex.Unlock()
				}

				msg.MsgType = Raw
				msg.Source = -1
				msg.Mask = common.EmptyMask
				msg.Signature.Clear()
				msgBytes := node.ConvertMsg.MessageToBytes(*msg)

				logrus.Debug("Signing message ", *msgBytes, " at node ", node.Id)
				signature := node.PrivateKey.Sign(*msgBytes)

				// Adding signature and ack to message. These fields were empty when message got signed
				msg.MsgType = Ack
				msg.Source = node.Id
				msg.Mask.SetBit(&common.EmptyMask, node.Id, 1)
				msg.Signature = signature
				msgBytes = node.ConvertMsg.MessageToBytes(*msg)
				node.Comm.Send(*msgBytes, msg.Source)

			case Catchup:
				mutex.Lock()
				if msg.Source == node.CurrentMsg.Source && msg.Step > nodeTimeStep {
					fmt.Printf("Catchup: node (%d,step %d), msg(source %d ,step %d)\n", node.Id, node.TimeStep, msg.Source, msg.Step)

					node.History = append(node.History, msg.History[nodeTimeStep:]...)
					node.Advance(msg.Step)

				}
				mutex.Unlock()
			}
		}(nodeTimeStep)

	}
	return
}
