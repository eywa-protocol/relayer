package uptime

import (
	"context"
	"time"

	"github.com/libp2p/go-flow-metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const ProtocolId protocol.ID = "/p2p/rpc/uptime"

const (
	RpcService           = "RpcApiUptime"
	RpcServiceFuncUptime = "Uptime"
	RpcServiceFuncReset  = "Reset"
)

type Service struct {
	leader         peer.ID
	uptimeRegistry *flow.MeterRegistry
}

type RpcApiUptime struct {
	server *Server
}

type UpTime struct {
	PeerId peer.ID
	Dur    time.Duration
}

type PearMsg struct {
	PeerId peer.ID
}

type UpList []UpTime

type Result struct {
	Uptimes UpList
}

func (r *RpcApiUptime) Uptime(_ context.Context, in PearMsg, out *Result) error {
	var err error
	*out, err = r.server.UptimeHandler(in.PeerId)
	return err
}

func (r *RpcApiUptime) Reset(_ context.Context, in PearMsg, out *PearMsg) error {
	*out = r.server.ResetHandler(in.PeerId)
	return nil
}
