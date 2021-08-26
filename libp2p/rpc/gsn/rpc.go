package gsn

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

const ProtocolId protocol.ID = "/p2p/rpc/gsn"

const (
	RpcService            = "RpcApiGsn"
	RpcServiceFuncExecute = "Execute"
)

type RpcForwarderForwardRequest struct {
	From  common.Address
	To    common.Address
	Value string
	Gas   string
	Nonce string
	Data  []byte
}

func NewRpcForwarderForwardRequest(req *wrappers.IForwarderForwardRequest) RpcForwarderForwardRequest {
	return RpcForwarderForwardRequest{
		From:  req.From,
		To:    req.To,
		Value: req.Value.String(),
		Gas:   req.Gas.String(),
		Nonce: req.Nonce.String(),
		Data:  req.Data,
	}
}

func (r RpcForwarderForwardRequest) DumpToIForwarderForwardRequest() (*wrappers.IForwarderForwardRequest, error) {

	dump := &wrappers.IForwarderForwardRequest{
		From: r.From,
		To:   r.To,
		Data: r.Data,
	}

	var ok bool
	if dump.Value, ok = new(big.Int).SetString(r.Value, 10); !ok {

		return nil, fmt.Errorf("can not parse request value [%s]",
			dump.Value)
	} else if dump.Gas, ok = new(big.Int).SetString(r.Gas, 10); !ok {

		return nil, fmt.Errorf("can not parse request gas [%s]",
			dump.Gas)
	} else if dump.Nonce, ok = new(big.Int).SetString(r.Nonce, 10); !ok {

		return nil, fmt.Errorf("can not parse request nonce [%s]",
			dump.Nonce)
	} else {
		return dump, nil
	}

}

type ExecuteRequest struct {
	ChainId         string
	ForwardRequest  RpcForwarderForwardRequest
	DomainSeparator common2.Bytes32
	RequestTypeHash common2.Bytes32
	SuffixData      []byte
	Signature       []byte
}

type ExecuteResult struct {
	ChainId string
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
		*out = *res
		return nil
	}

}
