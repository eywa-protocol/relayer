package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sync"

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
func (node *Session) verifyThresholdWitnesses(msg MessageWithSig) (err error) {
	if pop := Popcount(&msg.Mask); pop < node.ThresholdAck {
		err = fmt.Errorf("not enough sigantures: %d(%s) < %d", pop, msg.Mask.Text(16), node.ThresholdAck)
		return
	}

	// Verify message signature
	body, _ := json.Marshal(msg.Body)
	if !msg.Signature.VerifyMultisig(node.EpochPublicKey, msg.PublicKey, body, &msg.Mask) {
		return fmt.Errorf("threshold signature mismatch for '%s'", body)
	}
	logrus.Tracef("Aggregated Signature VERIFIED ! ! !")
	return
}

func (node *Session) verifyAckSignature(msg MessageWithSig) (err error) {
	body, _ := json.Marshal(msg.Body)
	mask := big.NewInt(0)
	mask.SetBit(mask, msg.Source, 1)
	if !msg.Signature.VerifyMultisig(node.EpochPublicKey, node.PublicKeys[msg.Source], body, mask) {
		err = fmt.Errorf("ACK signature mismatch for '%s'", body)
		return
	}
	return
}

// StartProtocol  will change the step of the node to a new one and then broadcast a message to the network.
func (node *Session) StartProtocol() {
	msg := MessageWithSig{
		Header:  Header{node.Id, Announce},
		Body:    Body{node.Topic.String()},
		History: make([]MessageWithSig, 0),
	}

	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.BroadcastTo(node.Topic, *msgBytes)
}

// WaitForProtocolMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Session) WaitForProtocolMsg(consensusAgreed chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithCancel(node.Ctx)
	defer cancel()
	mutex := &sync.Mutex{}
	for ctx.Err() == nil {
		// For now we assume that the underlying receive function is blocking
		rcvdMsg := node.Comm.ReceiveFrom(node.Topic.String(), ctx)
		if rcvdMsg == nil {
			logrus.Warn("protocol received empty message")
			break
		}

		wg.Add(1)
		go func(msgBytes *[]byte) {
			defer wg.Done()
			msg := node.ConvertMsg.BytesToModelMessage(*msgBytes)
			if msg.BridgeEventHash != node.Topic.String() {
				logrus.Warnf("NOT MY MESSAGE %s != %s", msg.BridgeEventHash, node.Topic.String())
				return
			}

			switch msg.MsgType {
			case Prepared:
				err := node.verifyThresholdWitnesses(*msg)
				if err != nil {
					logrus.Error("Prepared threshold failed: ", err, " at node ", node.Id)
					return
				}
				logrus.Debugf("Verified 'Prepared' Signature at node %d from node %d", node.Id, msg.Source)

				logrus.Tracef("Consensus achieved by node %v", node.Id)
				mutex.Lock()
				if ctx.Err() == nil {
					consensusAgreed <- true
					cancel()
				}
				mutex.Unlock()

			case Prepare:
				// Checking that the ack is for message of this step
				mutex.Lock()
				if (node.Acks >= node.ThresholdAck) || (node.SigMask.Bit(msg.Source) != 0) {
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
					node.Comm.BroadcastTo(node.Topic, *msgBytes)
				}

				mutex.Unlock()

			case Announce:
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
				node.Comm.BroadcastTo(node.Topic, *msgBytes)
			}
		}(rcvdMsg)
	}
	return
}
