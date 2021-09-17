package utils

import (
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/protocol"
	mh "github.com/multiformats/go-multihash"
)

func ProtocolToCid(id protocol.ID) (cid.Cid, error) {
	h, err := mh.Sum([]byte(id), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, h), nil
}
