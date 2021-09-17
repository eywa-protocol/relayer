package gsn

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/libp2p/go-libp2p-core/host"
	discovery "github.com/libp2p/go-libp2p-discovery"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p/utils"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

type ForwarderServerNode interface {
	GetDht() *dht.IpfsDHT
	GetOwner(chainId *big.Int) (*bind.TransactOpts, error)
	GetForwarder(chainId *big.Int) (*wrappers.Forwarder, error)
}

type Server struct {
	mx        *sync.Mutex
	node      ForwarderServerNode
	rpcServer *rpc.Server
}

// NewServer create new gsn forwarder rpc server and advertise it in network
func NewServer(ctx context.Context, host host.Host, node ForwarderServerNode) (*Server, error) {
	s := &Server{
		mx:        new(sync.Mutex),
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
	s.mx.Lock()
	defer s.mx.Unlock()
	logrus.Infof("request: %v", request)
	if chainId, ok := new(big.Int).SetString(request.ChainId, 10); !ok {

		return nil, fmt.Errorf("can not parse chain id [%s]",
			request.ChainId)
	} else if owner, err := s.node.GetOwner(chainId); err != nil {
		return nil, fmt.Errorf("can not get owner for chain id [%s] on error: %w",
			request.ChainId, err)
	} else if forwarder, err := s.node.GetForwarder(chainId); err != nil {
		return nil, fmt.Errorf("can not get forwarder for chain id [%s] on error: %w",
			request.ChainId, err)
	} else if forwardRequest, err := request.ForwardRequest.DumpToIForwarderForwardRequest(); err != nil {
		return nil, fmt.Errorf("can not dump forward request for chain id [%s] on error: %w",
			request.ChainId, err)

	} else if tx, err := forwarder.Execute(owner,
		*forwardRequest,
		request.DomainSeparator,
		request.RequestTypeHash,
		request.SuffixData,
		request.Signature); err != nil {
		return nil, fmt.Errorf("can not execute request for chain id [%s] on error:%w",
			request.ChainId, err)
	} else {
		return &ExecuteResult{
			ChainId: request.ChainId,
			TxHash:  tx.Hash(),
		}, nil
	}
}
