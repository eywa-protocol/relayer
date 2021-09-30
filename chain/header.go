package chain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"math/big"
)

// Header is the V3 block header.
// V3 block header is exactly the same
// we copy the code instead of embedded v2 header into v3
// when we do type checking in NewBodyForMatchingHeader
// the embedded structure will return v2 header type instead of v3 type
type Header struct {
	Fields headerFields
}

// EncodeRLP encodes the header Fields into RLP format.
func (h *Header) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &h.Fields)
}

// DecodeRLP decodes the given RLP decode stream into the header Fields.
func (h *Header) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(&h.Fields)
}

// NewHeader creates a new header object.
func NewHeader(e Epoch) *Header {
	return &Header{headerFields{
		Number: big.NewInt(1),
		Time:   big.NewInt(11111111),
		Epoch:  &e,
	}}
}

type headerFields struct {
	ParentHash          common.Hash `json:"parentHash"       gencodec:"required"`
	Root                common.Hash `json:"stateRoot"        gencodec:"required"`
	Number              *big.Int    `json:"number"           gencodec:"required"`
	Time                *big.Int    `json:"timestamp"        gencodec:"required"`
	Epoch               *Epoch      `json:"epoch"            gencodec:"required"`
	LastCommitSignature [96]byte    `json:"lastCommitSignature"  gencodec:"required"`
	LastCommitBitmap    []byte      `json:"lastCommitBitmap"     gencodec:"required"` // Contains which validator signed
}

// ParentHash is the header hash of the parent block.  For the genesis block
// which has no parent by definition, this field is zeroed out.
func (h *Header) ParentHash() common.Hash {
	return h.Fields.ParentHash
}

// SetParentHash sets the parent hash field.
func (h *Header) SetParentHash(newParentHash common.Hash) {
	h.Fields.ParentHash = newParentHash
}

// Root is the state (account) trie root hash.
func (h *Header) Root() common.Hash {
	return h.Fields.Root
}
