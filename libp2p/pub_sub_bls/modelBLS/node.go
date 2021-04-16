package modelBLS

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
)

// Node is the struct used for keeping everything related to a node in TLC.
type Node struct {
	Id           int                    // Id of the node
	TimeStep     int                    // Node's local time step
	ThresholdAck int                    // Threshold on number of messages
	ThresholdWit int                    // Threshold on number of witnessed messages
	Acks         int                    // Number of acknowledges
	Wits         int                    // Number of witnesses
	Comm         CommunicationInterface // interface for communicating with other nodes
	ConvertMsg   MessageInterface
	CurrentMsg   MessageWithSig   // Message which the node is waiting for acks
	History      []MessageWithSig // History of received messages by a node
	PublicKeys   []kyber.Point    // Public keys of all nodes
	Signatures   [][]byte
	SigMask      *sign.Mask
	PrivateKey   kyber.Scalar // Private key of the node
	Suite        *pairing.SuiteBn256
}

// CommunicationInterface is a interface used for communicating with transport layer.
type CommunicationInterface interface {
	Send([]byte, int) // Send a message to a specific node
	Broadcast([]byte) // Broadcast messages to other nodes
	Receive() *[]byte // Blocking receive
	Disconnect()      // Disconnect node
	Reconnect(string) // Reconnect node
}
