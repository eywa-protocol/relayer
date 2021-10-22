package model

import "context"

// Node is the struct used for keeping everything related to a node in TLC.
type Node struct {
	Ctx          context.Context
	Id           int                    // Id of the node
	TimeStep     int                    // Node's local time step
	ThresholdAck int                    // Threshold on number of messages
	ThresholdWit int                    // Threshold on number of witnessed messages
	Acks         int                    // Number of acknowledges
	Wits         int                    // Number of witnesses
	Comm         CommunicationInterface // interface for communicating with other nodes
	ConvertMsg   MessageInterface
	CurrentMsg   Message   // Message which the node is waiting for acks
	History      []Message // History of received messages by a node

}

// CommunicationInterface is a interface used for communicating with transport layer.
type CommunicationInterface interface {
	Send([]byte, int)                // Send a message to a specific node
	Broadcast([]byte)                // Broadcast messages to other nodes
	Receive(context.Context) *[]byte // Blocking receive
	Disconnect()                     // Disconnect node
	Reconnect(string)                // Reconnect node
	BroadcastMsg(message Message)
}
