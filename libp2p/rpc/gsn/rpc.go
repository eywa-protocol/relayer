package gsn

import (
	"context"
	"math/big"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/protocol"
	mh "github.com/multiformats/go-multihash"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

const ProtocolId protocol.ID = "/p2p/rpc/gsn"

const (
	RpcService            = "RpcApiGsn"
	RpcServiceFuncExecute = "Execute"
)

type ExecuteRequest struct {
	ChainId         *big.Int
	ForwardRequest  wrappers.IForwarderForwardRequest
	DomainSeparator common2.Bytes32
	RequestTypeHash common2.Bytes32
	SuffixData      []byte
	Signature       []byte
}

type ExecuteResult struct {
	ChainId *big.Int
	TxId    string
}

type RpcApiGsn struct {
	server *Server
}

func (r *RpcApiGsn) Execute(_ context.Context, in ExecuteRequest, out *ExecuteResult) error {
	if res, err := r.server.ExecuteRequestHandler(&in); err != nil {
		logrus.WithField("gsnRequest", in).Error(err)
		return err
	} else {
		out = res
		return nil
	}

}

func GetCid(chainId *big.Int) (cid.Cid, error) {
	var builder strings.Builder

	builder.WriteString(chainId.String())
	builder.WriteString(string(ProtocolId))

	h, err := mh.Sum([]byte(builder.String()), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, h), nil
}
