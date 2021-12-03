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

// verifyThresholdWitnesses verifies that it's really witnessed by majority of sessions by checking the signature and number of them
func (session *Session) verifyThresholdWitnesses(msg MessageWithSig) (err error) {
	if pop := Popcount(&msg.Mask); pop < session.ThresholdAck {
		err = fmt.Errorf("not enough sigantures: %d(%s) < %d", pop, msg.Mask.Text(16), session.ThresholdAck)
		return
	}

	// Verify message signature
	body, _ := json.Marshal(msg.Body)
	if !msg.Signature.VerifyMultisig(session.EpochPublicKey, msg.PublicKey, body, &msg.Mask) {
		return fmt.Errorf("threshold signature mismatch for '%s'", body)
	}
	logrus.Tracef("Aggregated Signature VERIFIED ! ! !")
	return
}

func (session *Session) verifyAckSignature(msg MessageWithSig) (err error) {
	body, _ := json.Marshal(msg.Body)
	mask := big.NewInt(0)
	mask.SetBit(mask, msg.Source, 1)
	if !msg.Signature.VerifyMultisig(session.EpochPublicKey, session.PublicKeys[msg.Source], body, mask) {
		err = fmt.Errorf("ACK signature mismatch for '%s'", body)
		return
	}
	return
}

// StartProtocol  will change the step of the session to a new one and then broadcast a message to the network.
func (session *Session) StartProtocol() {
	msg := MessageWithSig{
		Header:  Header{session.Id, Announce},
		Body:    Body{session.Topic.String()},
		History: make([]MessageWithSig, 0),
	}

	msgBytes := session.ConvertMsg.MessageToBytes(msg)
	session.Comm.BroadcastTo(session.Topic, *msgBytes)
}

// WaitForProtocolMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (session *Session) WaitForProtocolMsg(consensusAgreed chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithCancel(session.Ctx)
	defer cancel()
	mutex := &sync.Mutex{}
	for ctx.Err() == nil {
		// For now we assume that the underlying receive function is blocking
		rcvdMsg := session.Comm.ReceiveFrom(session.Topic.String(), ctx)
		if rcvdMsg == nil {
			logrus.Warn("protocol received empty message")
			break
		}

		wg.Add(1)
		go func(msgBytes *[]byte) {
			defer wg.Done()
			msg := session.ConvertMsg.BytesToModelMessage(*msgBytes)
			if msg.BridgeEventHash != session.Topic.String() {
				logrus.Warnf("NOT MY MESSAGE %s != %s", msg.BridgeEventHash, session.Topic.String())
				return
			}

			switch msg.MsgType {
			case Prepared:
				err := session.verifyThresholdWitnesses(*msg)
				if err != nil {
					logrus.Error("Prepared threshold failed: ", err, " at session ", session.Id)
					return
				}
				logrus.Debugf("Verified 'Prepared' Signature at session %d from session %d", session.Id, msg.Source)

				logrus.Tracef("Consensus achieved by session %v", session.Id)
				mutex.Lock()
				if ctx.Err() == nil {
					consensusAgreed <- true
					cancel()
				}
				mutex.Unlock()

			case Prepare:
				// Checking that the ack is for message of this step
				mutex.Lock()
				if (session.Acks >= session.ThresholdAck) || (session.SigMask.Bit(msg.Source) != 0) {
					mutex.Unlock()
					return
				}

				err := session.verifyAckSignature(*msg)
				if err != nil {
					logrus.Error(err, " at session ", session.Id, msg.Source, msg.Signature.Marshal())
					return
				}
				logrus.Debugf("Verified Prepare Signature at session %d from session %d", session.Id, msg.Source)

				session.SigMask.SetBit(&session.SigMask, msg.Source, 1)
				logrus.Debugf("Session SigMask Merged: %x", session.SigMask.Int64())

				// Count acks toward the threshold
				session.Acks += 1
				session.PartSignature = session.PartSignature.Aggregate(msg.Signature)
				session.PartPublicKey = session.PartPublicKey.Aggregate(session.PublicKeys[msg.Source])

				if session.Acks >= session.ThresholdAck {
					// Send witnessed message if the acks are more than threshold
					outmsg := MessageWithSig{
						Header:    Header{session.Id, Prepared},
						Body:      msg.Body,
						Mask:      session.SigMask,
						Signature: session.PartSignature,
						PublicKey: session.PartPublicKey,
					}

					// Verify before sending message to others
					if err := session.verifyThresholdWitnesses(outmsg); err != nil {
						logrus.Error("verifyThresholdWitnesses1 ", err, ", session: ", session.Id, ", mask: ", msg.Mask.Text(16))
						return
					}

					msgBytes := session.ConvertMsg.MessageToBytes(outmsg)
					session.Comm.BroadcastTo(session.Topic, *msgBytes)
				}

				mutex.Unlock()

			case Announce:
				body, _ := json.Marshal(msg.Body)
				signature := session.PrivateKey.Multisign(body, session.EpochPublicKey, session.MembershipKey)
				logrus.Debugf("Sign message '%s' at session %d", body, session.Id)

				// Adding signature and ack to message. These fields were empty when message got signed
				outmsg := MessageWithSig{
					Header:    Header{session.Id, Prepare},
					Body:      msg.Body,
					Signature: signature,
				}
				msgBytes = session.ConvertMsg.MessageToBytes(outmsg)
				session.Comm.BroadcastTo(session.Topic, *msgBytes)
			}
		}(rcvdMsg)
	}
	return
}
