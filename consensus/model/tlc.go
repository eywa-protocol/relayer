package model

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const ChanLen = 500

var Logger1 *log.Logger

// Advance will change the step of the node to a new one and then broadcast a message to the network.
func (node *Node) Advance(step int) {
	logFile, _ := os.OpenFile("../logs/tlc.log", os.O_RDWR|os.O_CREATE, 0666)
	Logger1 = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	node.TimeStep = step
	node.Acks = 0
	node.Wits = 0

	//fmt.Printf("node %d , Broadcast in timeStep %d,%#v\n", node.Id, node.TimeStep, node.History)
	Logger1.SetPrefix(strconv.FormatInt(time.Now().Unix(), 10) + " ")

	Logger1.Printf("NODE ID %d, TIME STEP %d\n", node.Id, node.TimeStep)

	msg := Message{
		Source:  node.Id,
		MsgType: Raw,
		Step:    node.TimeStep,
		History: make([]Message, 0),
	}
	node.CurrentMsg = msg
	msgBytes := node.ConvertMsg.MessageToBytes(msg)
	node.Comm.Broadcast(*msgBytes)
	Logger1.Printf("NODE ID %d, node.CurrentMsg %v\n", node.Id, node.CurrentMsg)

	//node.Comm.BroadcastMsg(node.CurrentMsg)
}

// waitForMsg waits for upcoming messages and then decides the next action with respect to msg's contents.
func (node *Node) WaitForMsg(stop int) {
	end := false
	msgChan := make(chan *[]byte, ChanLen)
	for node.TimeStep <= stop && !end {
		// For now we assume that the underlying receive function is blocking
		rcvdMsg := node.Comm.Receive(node.Ctx)
		if rcvdMsg == nil {
			continue
		}
		msgChan <- rcvdMsg

		go func() {
			msgBytes := <-msgChan

			msg := node.ConvertMsg.BytesToModelMessage(*msgBytes)

			//fmt.Printf("node %d in step %d ;Received MSG with step %d type %d source: %d\n", node.Id, node.TimeStep, msg.Step, msg.MsgType, msg.Source)

			// Used for stopping the execution after some timesteps
			if node.TimeStep == stop {
				fmt.Println("Break reached by node ", node.Id)
				end = true
				return
			}

			// If the received message is from a lower step, send history to the node to catch up
			if msg.Step < node.TimeStep {
				if msg.MsgType == Raw {
					msg.MsgType = Catchup
					msg.Step = node.TimeStep
					msg.History = node.History
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)
				}
				return
			}

			switch msg.MsgType {
			case Wit:
				if msg.Step > node.TimeStep+1 {
					return
				} else if msg.Step == node.TimeStep+1 { // Node needs to catch up with the message
					// Update nodes local history. Append history from message to local history
					node.History = append(node.History, *msg)

					// Advance
					node.Advance(msg.Step)
					node.Wits += 1

				} else if msg.Step == node.TimeStep {
					// Count message toward the threshold
					node.Wits += 1
					//fmt.Printf("WITS: node %d , %d\n", node.Id, node.Wits)

					if node.Wits >= node.ThresholdWit {
						fmt.Printf("Witnessed by threshold. node %d !\n", node.Id)
						// Log the message in history
						node.History = append(node.History, *msg)
						// Advance to next time step
						node.Advance(node.TimeStep + 1)
					}
				}

			case Ack:
				//fmt.Printf("received ACK. node %d %d\n", node.Id, msg.Source)

				// Checking that the ack is for message of this step
				if (msg.Source != node.CurrentMsg.Source) || (msg.Step != node.CurrentMsg.Step) {
					return
				}
				//fmt.Printf("received ACK. node %d !\n", node.Id)

				// Count acks toward the threshold
				node.Acks += 1

				if node.Acks >= node.ThresholdAck {
					// Send witnessed message if the acks are more than threshold
					fmt.Printf("Acked by threshold. node %d !\n", node.Id)
					msg.MsgType = Wit
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Broadcast(*msgBytes)
				}

			case Raw:
				if msg.Step > node.TimeStep+1 {
					return
				} else if msg.Step == node.TimeStep+1 { // Node needs to catch up with the message
					// Update nodes local history. Append history from message to local history
					node.History = append(node.History, *msg)

					// Advance
					node.Advance(msg.Step)

					// ALso send ack for the received message
					msg.MsgType = Ack
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Send(*msgBytes, msg.Source)

				} else if msg.Step == node.TimeStep {
					msg.MsgType = Ack
					//fmt.Printf("ACKing by node %d, for msg %d\n", node.Id, msg.Source)
					msgBytes := node.ConvertMsg.MessageToBytes(*msg)
					node.Comm.Send(*msgBytes, msg.Source)
				}

			case Catchup:
				if msg.Source == node.CurrentMsg.Source && msg.Step > node.TimeStep {
					fmt.Printf("Catchup: node (%d,step %d), msg(source %d ,step %d)\n", node.Id, node.TimeStep, msg.Source, msg.Step)
					node.History = append(node.History, msg.History[node.TimeStep:]...)
					node.Advance(msg.Step)
				}
			}
		}()

	}
}
