package _interface

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Header defines the block header interface.
type Header interface {
	// ParentHash is the header hash of the parent block.  For the genesis block
	// which has no parent by definition, this field is zeroed out.
	ParentHash() common.Hash

	// SetParentHash sets the parent hash field.
	SetParentHash(newParentHash common.Hash)

	// Root is the state (account) trie root hash.
	Root() common.Hash

	// SetRoot sets the state trie root hash field.
	SetRoot(newRoot common.Hash)

	// TxHash is the transaction trie root hash.
	TxHash() common.Hash

	// SetTxHash sets the transaction trie root hash field.
	SetTxHash(newTxHash common.Hash)

	// Number is the block number.
	//
	// The returned instance is a copy; the caller may do anything with it.
	Number() *big.Int

	// SetNumber sets the block number.
	//
	// It stores a copy; the caller may freely modify the original.
	SetNumber(newNumber *big.Int)

	// Time is the UNIX timestamp of this block.
	//
	// The returned instance is a copy; the caller may do anything with it.
	Time() *big.Int

	// SetTime sets the UNIX timestamp of this block.
	//
	// It stores a copy; the caller may freely modify the original.
	SetTime(newTime *big.Int)

	// Extra is the extra data field of this block.
	//
	// The returned slice is a copy; the caller may do anything with it.
	Extra() []byte

	// SetExtra sets the extra data field of this block.
	//
	// It stores a copy; the caller may freely modify the original.
	SetExtra(newExtra []byte)

	// MixDigest is the mixhash.
	//
	// This field is a remnant from Ethereum, and Harmony does not use it and always
	// zeroes it out.
	MixDigest() common.Hash

	// SetMixDigest sets the mixhash of this block.
	SetMixDigest(newMixDigest common.Hash)

	// ViewID is the ID of the view in which this block was originally proposed.
	//
	// It normally increases by one for each subsequent block, or by more than one
	// if one or more PBFT/FBFT view changes have occurred.
	//
	// The returned instance is a copy; the caller may do anything with it.
	ViewID() *big.Int

	// SetViewID sets the view ID in which the block was originally proposed.
	//
	// It stores a copy; the caller may freely modify the original.
	SetViewID(newViewID *big.Int)

	// Epoch is the epoch number of this block.
	//
	// The returned instance is a copy; the caller may do anything with it.
	Epoch() *big.Int

	// SetEpoch sets the epoch number of this block.
	//
	// It stores a copy; the caller may freely modify the original.
	SetEpoch(newEpoch *big.Int)

	// LastCommitSignature is the FBFT commit group signature for the last block.
	LastCommitSignature() [96]byte

	// SetLastCommitSignature sets the FBFT commit group signature for the last
	// block.
	SetLastCommitSignature(newLastCommitSignature [96]byte)

	// LastCommitBitmap is the signatory bitmap of the previous block.  Bit
	// positions index into committee member array.
	//
	// The returned slice is a copy; the caller may do anything with it.
	LastCommitBitmap() []byte

	// SetLastCommitBitmap sets the signatory bitmap of the previous block.
	//
	// It stores a copy; the caller may freely modify the original.
	SetLastCommitBitmap(newLastCommitBitmap []byte)

	// Hash returns the block hash of the header, which is simply the legacy
	// Keccak256 hash of its RLP encoding.
	Hash() common.Hash
}
