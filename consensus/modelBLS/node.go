package modelBLS

import (
	"math/big"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/libp2p/go-libp2p-core/peer"
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
	CurrentMsg        MessageWithSig   // Message which the node is waiting for acks
	History           []MessageWithSig // History of received messages by a node
	PublicKeys        []bls.PublicKey  // Public keys of all nodes
	PartSignature     bls.Signature    // Aggregated partial signature collected by this node in this step
	PartPublicKey     bls.PublicKey    // Aggregated partial public key collected by this node in this step
	SigMask           big.Int          // Bitmask of nodes that the right signature is received from
	PrivateKey        bls.PrivateKey   // Private key of the node
	EpochPublicKey    bls.PublicKey    // Arrtegated public key of all participants of the current epoch
	MembershipKey     bls.Signature    // Membership key of this node
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
