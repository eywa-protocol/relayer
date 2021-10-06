package modelBLS

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

// Node is the struct used for keeping everything related to a node in TLC.
type Node struct {
	Id                int                    // Id of the node
	TimeStep          int                    // Node's local time step
	ThresholdAck      int                    // Threshold on number of messages
	ThresholdWit      int                    // Threshold on number of witnessed messages
	Acks              int                    // Number of acknowledges
	Wits              int                    // Number of witnesses
	Comm              CommunicationInterface // interface for communicating with other nodes
	ConvertMsg        MessageInterface
	CurrentMsg        MessageWithSig        // Message which the node is waiting for acks
	History           []MessageWithSig      // History of received messages by a node
	PublicKeys        []common.BlsPublicKey // Public keys of all nodes
	Signatures        []common.BlsSignature
	SigMask           big.Int
	PrivateKey        common.BlsPrivateKey // Private key of the node
	Participants      []peer.ID
	CurrentRendezvous string
	Leader            peer.ID
}

// CommunicationInterface is a interface used for communicating with transport layer.
type CommunicationInterface interface {
	Send([]byte, int) // Send a message to a specific node
	Broadcast([]byte) // Broadcast messages to other nodes
	Receive() *[]byte // Blocking receive
	Disconnect()      // Disconnect node
	Reconnect(string) // Reconnect node
}
