package uptime

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-flow-metrics"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/config"
)

type Client struct {
	Service
	rpcClient *rpc.Client
	pIds      peer.IDSlice
}

func NewLeader(host host.Host, selfRegistry *flow.MeterRegistry, peerIds peer.IDSlice) *Client {
	filteredIds := make(peer.IDSlice, 0, len(peerIds))
	for i := range peerIds {
		if peerIds[i] != host.ID() {
			filteredIds = append(filteredIds, peerIds[i])
		}
	}
	return &Client{
		Service: Service{
			leader:         host.ID(),
			uptimeRegistry: selfRegistry,
		},
		rpcClient: rpc.NewClient(host, ProtocolId),
		pIds:      filteredIds,
	}
}

func (c *Client) Uptime() UpList {

	res := make([]*Result, len(c.pIds))
	errs := c.rpcClient.MultiCall(
		c.ctx(),
		c.pIds,
		RpcService,
		RpcServiceFuncUptime,
		PearMsg{
			PeerId: c.leader,
		},
		c.upTimeReplies(res),
	)

	uptimes := make(map[peer.ID][]time.Duration, len(c.pIds)+1)

	c.uptimeRegistry.ForEach(func(s string, meter *flow.Meter) {
		if pId, err := peer.Decode(s); err != nil {
			logrus.WithField("peerId", s).Error(fmt.Errorf("decode peer id error: %w", err))
		} else {
			snapshot := meter.Snapshot()
			if _, ok := uptimes[pId]; !ok {
				uptimes[pId] = make([]time.Duration, 0, len(c.pIds)+1)
			}
			uptimes[pId] = append(uptimes[pId], time.Duration(snapshot.Total)*time.Second)
		}

	})

	for i, err := range errs {
		if err != nil {
			logrus.WithField("peerId", c.pIds[i].Pretty()).Error(fmt.Errorf("can not get peer uptime on error: %w", err))
		} else {
			for _, uptime := range res[i].Uptimes {
				if _, ok := uptimes[uptime.PeerId]; !ok {
					uptimes[uptime.PeerId] = make([]time.Duration, 0, len(c.pIds)+1)
				}
				uptimes[uptime.PeerId] = append(uptimes[uptime.PeerId], uptime.Dur)
			}
		}
	}

	resList := make(UpList, 0, len(uptimes))
	for id, durations := range uptimes {
		cnt := 0
		var dur time.Duration
		for _, duration := range durations {
			if duration >= config.App.UptimeReportInterval-config.App.TickerInterval &&
				duration <= config.App.UptimeReportInterval+config.App.TickerInterval {
				cnt++
				if duration > dur {
					dur = duration
				}
			}
		}
		if cnt > len(durations)/2 {
			resList = append(resList, UpTime{
				PeerId: id,
				Dur:    dur,
			})
		}
	}

	return resList
}

func (c *Client) Reset() {

	res := make([]*PearMsg, len(c.pIds))
	errs := c.rpcClient.MultiCall(
		c.ctx(),
		c.pIds,
		RpcService,
		RpcServiceFuncReset,
		PearMsg{
			PeerId: c.leader,
		},
		c.resetReplies(res),
	)
	for i, err := range errs {
		if err != nil {
			logrus.WithField("peerId", c.pIds[i].Pretty()).Error(fmt.Errorf("can not reset peer uptime on error: %w", err))
		} else {
			logrus.Debugf("uptime metrics reseted for pearId: %s", res[i].PeerId.Pretty())
		}
	}
	c.uptimeRegistry.Clear()
	logrus.Debugf("uptime metrics reseted for pearId: %s", c.leader.Pretty())
}

func (c *Client) ctx() []context.Context {
	l := len(c.pIds)
	res := make([]context.Context, 0, l)
	for i := 0; i < l; i++ {
		res = append(res, context.Background())
	}
	return res
}

func (c *Client) upTimeReplies(resList []*Result) []interface{} {
	res := make([]interface{}, len(resList))
	for i := range resList {
		resList[i] = &Result{Uptimes: UpList{}}
		res[i] = resList[i]
	}
	return res
}

func (c *Client) resetReplies(peerMsgs []*PearMsg) []interface{} {
	res := make([]interface{}, len(peerMsgs))
	for i := range peerMsgs {
		peerMsgs[i] = &PearMsg{}
		res[i] = peerMsgs[i]
	}
	return res
}
