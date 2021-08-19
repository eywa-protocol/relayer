package gsn

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/libp2p/go-libp2p-core/host"
	discovery "github.com/libp2p/go-libp2p-discovery"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/utils"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type ForwarderServerNode interface {
	GetDht() *dht.IpfsDHT
	GetOwner(chainId *big.Int) (*bind.TransactOpts, error)
	GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error)
}

type Server struct {
	node      ForwarderServerNode
	rpcServer *rpc.Server
}

// NewServer create new gsn forwarder rpc server and advertise it in network
func NewServer(ctx context.Context, host host.Host, node ForwarderServerNode) (*Server, error) {
	s := &Server{
		node:      node,
		rpcServer: rpc.NewServer(host, ProtocolId),
	}

	rpcServer := RpcApiGsn{server: s}

	routingDiscovery := discovery.NewRoutingDiscovery(s.node.GetDht())

	if err := s.rpcServer.Register(&rpcServer); err != nil {

		return nil, err
	} else if cid, err := utils.ProtocolToCid(ProtocolId); err != nil {

		return nil, err
	} else if err := routingDiscovery.Provide(ctx, cid, true); err != nil {

		return nil, err
	} else {

		return s, nil
	}
}

func (s *Server) ExecuteRequestHandler(request *ExecuteRequest) (*ExecuteResult, error) {

	if owner, err := s.node.GetOwner(request.ChainId); err != nil {
		return nil, fmt.Errorf("can not get owner for chain id [%s] on error: %w",
			request.ChainId.String(), err)
	} else if forwarder, err := s.node.GetForwarder(request.ChainId); err != nil {
		return nil, fmt.Errorf("can not get forwarder for chain id [%s] on error: %w",
			request.ChainId.String(), err)
	} else if tx, err := forwarder.Execute(owner,
		request.ForwardRequest,
		request.DomainSeparator,
		request.RequestTypeHash,
		request.SuffixData,
		request.Signature); err != nil {
		return nil, fmt.Errorf("can not execute request for chain id [%s] on error:%w",
			request.ChainId.String(), err)
	} else {
		return &ExecuteResult{
			ChainId: request.ChainId,
			TxId:    tx.Hash().Hex(),
		}, nil
	}
}
