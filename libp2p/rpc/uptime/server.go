package uptime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-flow-metrics"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/sirupsen/logrus"
)

var ErrLeaderNotMatch = errors.New("request leader not match to node leader")

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	Service
	rpcServer *rpc.Server
}

func (s *Server) UptimeHandler(leader peer.ID) (Result, error) {
	res := make(UpList, 0)
	if s.leader == leader {
		logrus.Tracef("collect uptime on peerId: %s", s.rpcServer.ID())
		s.uptimeRegistry.ForEach(func(s string, meter *flow.Meter) {
			if peerId, err := peer.Decode(s); err != nil {
				err = fmt.Errorf("decode peer Id error: %w", err)
				logrus.WithField("peer_id_string", s).Error(err)
			} else {
				snapshot := meter.Snapshot()
				res = append(res, UpTime{
					PeerId: peerId,
					Dur:    time.Duration(snapshot.Total) * time.Second,
				})
			}

		})
	} else {
		logrus.WithFields(logrus.Fields{
			"request_leader": leader,
			"node_leader":    s.leader,
		}).Error(ErrLeaderNotMatch)
		return Result{}, ErrLeaderNotMatch
	}
	return Result{
		Uptimes: res,
	}, nil
}

func (s *Server) ResetHandler(leader peer.ID) PearMsg {
	if s.leader == leader {
		logrus.Tracef("clear uptime registry on peerId: %s", s.rpcServer.ID())
		s.uptimeRegistry.Clear()
		defer s.cancel()
	}
	return PearMsg{PeerId: s.rpcServer.ID()}
}

func (s *Server) WaitForReset() {
	<-s.ctx.Done()
}

func NewServer(host host.Host, leader peer.ID, registry *flow.MeterRegistry) (*Server, error) {
	s := &Server{
		Service: Service{
			uptimeRegistry: registry,
			leader:         leader,
		},
		rpcServer: rpc.NewServer(host, ProtocolId),
	}
	s.ctx, s.cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Minute))

	upTimeServer := RpcApiUptime{server: s}
	err := s.rpcServer.Register(&upTimeServer)
	if err != nil {
		return nil, err
	}
	return s, nil
}
