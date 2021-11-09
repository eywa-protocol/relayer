package model

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/sirupsen/logrus"
)

const ChanLen = 500

var Logger1 *log.Logger

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

// verifyThresholdWitnesses verifies that it's really witnessed by majority of nodes by checking the signature and number of them
func (node *Node) verifyThresholdWitnesses(msg MessageWithSig) (err error) {
	if pop := Popcount(&msg.Mask); pop < node.ThresholdAck {
		err = fmt.Errorf("Not enough sigantures: %d(%s) < %d", pop, msg.Mask.Text(16), node.ThresholdAck)
		return
	}

	// Verify message signature
	body, _ := json.Marshal(msg.Body)
	if !msg.Signature.VerifyMultisig(node.EpochPublicKey, msg.PublicKey, body, &msg.Mask) {
		return fmt.Errorf("Threshold signature mismatch for '%s'", body)
	}
	logrus.Tracef("Aggregated Signature VERIFIED ! ! !")
	return
}

func (node *Node) verifyAckSignature(msg MessageWithSig) (err error) {
	body, _ := json.Marshal(msg.Body)
	mask := big.NewInt(0)
	mask.SetBit(mask, msg.Source, 1)
	if !msg.Signature.VerifyMultisig(node.EpochPublicKey, node.PublicKeys[msg.Source], body, mask) {
		err = fmt.Errorf("ACK signature mismatch for '%s'", body)
		return
	}
	return
}

// AdvanceStep  will change the step of the node to a new one and then broadcast a message to the network.
func (node *Node) AdvanceStep(step int) {
	node.TimeStep = step
	node.Acks = 0
	node.Wits = 0

	// fmt.Printf("Node ID %d, STEP %d\n", node.Id, node.TimeStep)

	msg := MessageWithSig{
		Header:  Header{node.Id, Announce},
		Body:    Body{node.TimeStep, node.Topic.String()},
		History: make([]MessageWithSig, 0),
	}

	node.CurrentMsg = msg
	node.PartSignature = bls.ZeroSignature()
	node.PartPublicKey = bls.ZeroPublicKey()
	node.SigMask = bls.EmptyMultisigMask()

	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.Broadcast(*msgBytes)
}

func (node *Node) DisconnectPubSub() {
	node.Comm.Disconnect()

}

// WaitForProtocolMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForProtocolMsg(consensusAgreed chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithCancel(node.Ctx)
	defer cancel()
	mutex := &sync.Mutex{}
	msgChan := make(chan *[]byte, ChanLen)
	stop := 1
	for ctx.Err() == nil {
		// For now we assume that the underlying receive function is blocking
		rcvdMsg := node.Comm.Receive(ctx)
		if rcvdMsg == nil {
			logrus.Warn("rcvdMsg:", rcvdMsg)
			break
		}

		msg := node.ConvertMsg.BytesToModelMessage(*rcvdMsg)
		if msg.BridgeEventHash == node.Topic.String() {
			msgChan <- rcvdMsg
			wg.Add(1)
			go func() {
				defer wg.Done()
				msgBytes := <-msgChan
				msg := node.ConvertMsg.BytesToModelMessage(*msgBytes)

				mutex.Lock()
				nodeTimeStep := node.TimeStep
				mutex.Unlock()

				if nodeTimeStep == stop {
					logrus.Infof(" Consensus achieved by node %v", node.Id)
					mutex.Lock()
					if ctx.Err() == nil {
						consensusAgreed <- true
						cancel()
					}
					mutex.Unlock()
					return
				}

				// If the received message is from a lower step, send history to the node to catch up
				if msg.Step < nodeTimeStep {
					if msg.MsgType == Announce {
						msg.MsgType = Commit
						msg.Step = nodeTimeStep
						msg.History = node.History
						msgBytes := node.ConvertMsg.MessageToBytes(*msg)
						node.Comm.Broadcast(*msgBytes)

					}
					return
				}

				switch msg.MsgType {
				case Prepared:
					if msg.Step > nodeTimeStep+1 {
						return
					}

					err := node.verifyThresholdWitnesses(*msg)
					if err != nil {
						logrus.Error("Vit threshold failed: ", err, " at node ", node.Id, msg.Signature.Marshal())
						return
					}
					logrus.Debugf("Verified Vit Signature at node %d from node %d", node.Id, msg.Source)

					mutex.Lock()
					node.Wits += 1
					node.AdvanceStep(nodeTimeStep + 1)
					mutex.Unlock()

				case Prepare:
					// Checking that the ack is for message of this step
					mutex.Lock()
					if (msg.Step != node.CurrentMsg.Step) || (node.Acks >= node.ThresholdAck) || (node.SigMask.Bit(msg.Source) != 0) {
						mutex.Unlock()
						return
					}

					err := node.verifyAckSignature(*msg)
					if err != nil {
						logrus.Error(err, " at node ", node.Id, msg.Source, msg.Signature.Marshal())
						return
					}
					logrus.Debugf("Verified Prepare Signature at node %d from node %d", node.Id, msg.Source)

					node.SigMask.SetBit(&node.SigMask, msg.Source, 1)
					logrus.Debugf("Node SigMask Merged: %x", node.SigMask.Int64())

					// Count acks toward the threshold
					node.Acks += 1
					node.PartSignature = node.PartSignature.Aggregate(msg.Signature)
					node.PartPublicKey = node.PartPublicKey.Aggregate(node.PublicKeys[msg.Source])

					if node.Acks >= node.ThresholdAck {
						// Send witnessed message if the acks are more than threshold
						outmsg := MessageWithSig{
							Header:    Header{node.Id, Prepared},
							Body:      msg.Body,
							Mask:      node.SigMask,
							Signature: node.PartSignature,
							PublicKey: node.PartPublicKey,
						}

						// Verify before sending message to others
						if err := node.verifyThresholdWitnesses(outmsg); err != nil {
							logrus.Error("verifyThresholdWitnesses1 ", err, ", node: ", node.Id, ", mask: ", msg.Mask.Text(16))
							return
						}

						msgBytes := node.ConvertMsg.MessageToBytes(outmsg)
						node.Comm.Broadcast(*msgBytes)
					}

					mutex.Unlock()

				case Announce:
					if msg.Step > nodeTimeStep+1 {
						return
					} else if msg.Step == nodeTimeStep+1 { // Node needs to catch up with the message
						// Update nodes local history. Append history from message to local history
						mutex.Lock()
						node.History = append(node.History, *msg)

						// Advance
						node.AdvanceStep(msg.Step)
						mutex.Unlock()
					}

					body, _ := json.Marshal(msg.Body)
					signature := node.PrivateKey.Multisign(body, node.EpochPublicKey, node.MembershipKey)
					logrus.Debugf("Sign message '%s' at node %d", body, node.Id)

					// Adding signature and ack to message. These fields were empty when message got signed
					outmsg := MessageWithSig{
						Header:    Header{node.Id, Prepare},
						Body:      msg.Body,
						Signature: signature,
					}
					msgBytes = node.ConvertMsg.MessageToBytes(outmsg)
					node.Comm.Broadcast(*msgBytes)

				case Commit:
					mutex.Lock()
					if msg.Source == node.CurrentMsg.Source && msg.Step > nodeTimeStep {
						fmt.Printf("Commit: node (%d,step %d), msg(source %d ,step %d)\n", node.Id, node.TimeStep, msg.Source, msg.Step)

						node.History = append(node.History, msg.History[nodeTimeStep:]...)
						node.AdvanceStep(msg.Step)

					}
					mutex.Unlock()
				}
			}()
		} else {
			logrus.Warnf("NOT MY MESSAGE !!!!!!!!!!!!!!!!!!!  %s != %s", msg.BridgeEventHash, node.Topic.String())
			//node.DisconnectPubSub()
			continue
		}
	}
	return
}
