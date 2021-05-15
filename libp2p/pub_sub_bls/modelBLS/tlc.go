package modelBLS

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"log"
	"sync"
)

const ChanLen = 500

var Logger1 *log.Logger

// Advance will change the step of the node to a new one and then broadcast a message to the network.
func (node *Node) Advance(step int) {
	node.TimeStep = step
	node.Acks = 0
	node.Wits = 0

	fmt.Printf("node %d , Broadcast in timeStep %d\n", node.Id, node.TimeStep)
	fmt.Printf("Node ID %d, STEP %d\n", node.Id, node.TimeStep)

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

// waitForMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForMsg(stop int) (err error) {
	mutex := &sync.Mutex{}
	end := false
	msgChan := make(chan *[]byte, ChanLen)
	nodeTimeStep := 0

	for nodeTimeStep <= stop {
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

			logrus.Printf("node %d\n in nodeTimeStep %d;\nReceived MSG with\n msg.Step %d\n MsgType %d source: %d\n", node.Id, nodeTimeStep, msg.Step, msg.MsgType, msg.Source)

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

				err := node.verifyThresholdWitnesses(msg)
				if err != nil {
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
				if (msg.Source != node.CurrentMsg.Source) || (msg.Step != node.CurrentMsg.Step) || (node.Acks >= node.ThresholdAck) {
					mutex.Unlock()
					return
				}
				mutex.Unlock()
				fmt.Printf("node %d received ACK from node %d\n", node.Id, msg.Source)

				msgHash := calculateHash(*msg, node.ConvertMsg)

				err := node.verifyAckSignature(msg, msgHash)
				if err != nil {
					return
				}

				// add message's mask to existing mask
				mutex.Lock()
				err = node.SigMask.Merge(msg.Mask)
				if err != nil {
					logrus.Error(err)
					return
				}

				// Count acks toward the threshold
				node.Acks += 1

				keyMask, _ := sign.NewMask(node.Suite, node.PublicKeys, nil)
				err = keyMask.SetMask(msg.Mask)
				if err != nil {
					logrus.Error(err)
					panic(err)
				}
				index := keyMask.IndexOfNthEnabled(0)
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
						logrus.Println("node ", node.Id, "PANIC AggregateSignatures: ", node.Signatures, "Pub :", node.PublicKeys, "mask :", msg.Mask)
						panic(err)
					}
					msg.Signature, err = aggSignature.MarshalBinary()
					if err != nil {
						panic(err)
					}

					aggPubKey, err := bdn.AggregatePublicKeys(node.Suite, node.SigMask)

					// Verify before sending message to others
					err = bdn.Verify(node.Suite, aggPubKey, msgHash, msg.Signature)
					if err != nil {
						fmt.Println("node ", node.Id, "PANIC Sig: ", node.Signatures, "Pub :", node.PublicKeys, "mask :", msg.Mask)
						//panic(err)
						return
					}

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

				// Node has to sign message hash
				h := sha256.New()
				h.Write(*msgBytes)
				msgHash := h.Sum(nil)

				signature, err := bdn.Sign(node.Suite, node.PrivateKey, msgHash)
				if err != nil {
					panic(err)
				}

				// Adding signature and ack to message. These fields were empty when message got signed
				msg.Signature = signature

				// Add mask for the signature
				keyMask, _ := sign.NewMask(node.Suite, node.PublicKeys, nil)
				err = keyMask.SetBit(node.Id, true)
				if err != nil {
					panic(err)
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
	return err
}

func (node *Node) verifyThresholdWitnesses(msg *MessageWithSig) (err error) {
	// Verify that it's really witnessed by majority of nodes by checking the signature and number of them
	sig := msg.Signature
	mask := msg.Mask

	msg.Signature = nil
	msg.Mask = nil
	msg.MsgType = Raw

	h := sha256.New()
	h.Write(*node.ConvertMsg.MessageToBytes(*msg))
	msgHash := h.Sum(nil)

	keyMask, err := sign.NewMask(node.Suite, node.PublicKeys, nil)
	err = keyMask.SetMask(mask)
	if err != nil {
		return
	}

	if keyMask.CountEnabled() < node.ThresholdAck {
		err = errors.New("not Enough sigantures")
		return
	}

	aggPubKey, err := bdn.AggregatePublicKeys(node.Suite, keyMask)
	if err != nil {
		panic(err)
		return
	}

	// Verify message signature
	fmt.Println("RCVD AggSig: ", sig, "RCVD AggPub :", aggPubKey, "RCVD Hash :", msgHash)
	err = bdn.Verify(node.Suite, aggPubKey, msgHash, sig)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<< Aggregated Signature VERIFIED ! ! ! >>>>>>>>>>>>>>>>>>>>>>>>>>>>")

	return nil
}

func (node *Node) verifyAckSignature(msg *MessageWithSig, msgHash []byte) (err error) {

	keyMask, err := sign.NewMask(node.Suite, node.PublicKeys, nil)
	if err != nil {
		return
	}
	err = keyMask.SetMask(msg.Mask)
	if err != nil {
		logrus.Error(err)
		return
	}

	PubKey := append(node.PublicKeys)[keyMask.IndexOfNthEnabled(0)]

	err = bdn.Verify(node.Suite, PubKey, msgHash, msg.Signature)
	if err != nil {
		return
	}
	return
}

func calculateHash(msg MessageWithSig, converter MessageInterface) []byte {
	msg.Signature = nil
	msg.Mask = nil
	msg.MsgType = Raw

	h := sha256.New()
	h.Write(*converter.MessageToBytes(msg))
	msgHash := h.Sum(nil)
	return msgHash
}
