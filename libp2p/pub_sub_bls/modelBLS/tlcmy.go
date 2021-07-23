package modelBLS

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

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
		node.Signatures[i] = nil
	}
	mask, _ := sign.NewMask(node.Suite, node.PublicKeys, nil)
	node.SigMask = mask

	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.Broadcast(*msgBytes)
}

func (node *Node) DisconnectPubSub() {
	node.Comm.Disconnect()

}

// WaitForMsgNEW  waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForMsgNEW(consensusAgreed chan bool, wg *sync.WaitGroup) {
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

				err := node.verifyThresholdWitnesses(msg)
				if err != nil {
					return
				}
				mutex.Lock()
				node.Wits += 1
				node.TimeStep += 1
				node.Advance(nodeTimeStep + 1)
				mutex.Unlock()

			case Ack:
				// Checking that the ack is for message of this step
				mutex.Lock()
				if (msg.Source != node.CurrentMsg.Source) || (msg.Step != node.CurrentMsg.Step) || (node.Acks >= node.ThresholdAck) {
					mutex.Unlock()
					return
				}
				mutex.Unlock()
				msgHash := calculateHash(*msg, node.ConvertMsg)
				err := node.verifyAckSignature(msg, msgHash)
				if err != nil {
					logrus.Error(err)
				}
				//fmt.Print("verified Ack Signature\n")
				mutex.Lock()
				err = node.SigMask.Merge(msg.Mask)
				if err != nil {
					logrus.Error(err)
				}
				//fmt.Print("node SigMask Merged\n")
				// Count acks toward the threshold
				node.Acks += 1

				keyMask, _ := sign.NewMask(node.Suite, node.PublicKeys, nil)
				err = keyMask.SetMask(msg.Mask)
				if err != nil {
					logrus.Errorf(err.Error())
				}
				index := keyMask.IndexOfNthEnabled(0)
				// Add signature to the list of signatures
				if index == -1 {
					logrus.Error("no such pubkey")
					mutex.Unlock()
					break
				}
				node.Signatures[index] = msg.Signature
				if node.Acks >= node.ThresholdAck {
					// Send witnessed message if the acks are more than threshold
					msg.MsgType = Wit

					// Add aggregate signatures to message
					msg.Mask = node.SigMask.Mask()

					sigs := make([][]byte, 0)
					for _, sig := range node.Signatures {
						if sig != nil {
							sigs = append(sigs, sig)
						}
					}

					aggSignature, err := bdn.AggregateSignatures(node.Suite, sigs, node.SigMask)
					if err != nil {
						logrus.Error(err)
					}

					msg.Signature, err = aggSignature.MarshalBinary()
					if err != nil {
						logrus.Error(err)
					}
					aggPubKey, err := bdn.AggregatePublicKeys(node.Suite, node.SigMask)

					err = bdn.Verify(node.Suite, aggPubKey, msgHash, msg.Signature)
					if err != nil {
						logrus.Error(err)
					}

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

				// Node has to sign message hash
				h := sha256.New()
				h.Write(*msgBytes)
				msgHash := h.Sum(nil)

				signature, err := bdn.Sign(node.Suite, node.PrivateKey, msgHash)
				if err != nil {
					logrus.Error(err)
				}

				// Adding signature and ack to message. These fields were empty when message got signed
				msg.Signature = signature

				// Add mask for the signature
				keyMask, _ := sign.NewMask(node.Suite, node.PublicKeys, nil)
				err = keyMask.SetBit(node.Id, true)
				if err != nil {
					logrus.Error(err)
				}
				msg.Mask = keyMask.Mask()

				// Send ack for the received message
				msg.MsgType = Ack
				msgBytes := node.ConvertMsg.MessageToBytes(*msg)
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
