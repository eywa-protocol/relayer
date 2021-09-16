package bridge

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/prom/base"
)

const (
	namespace = "eywa"
	subsystem = "bridge"
)

type Metrics struct {
	base.MetricsServer
	disabled              bool
	reqReceivedCounter    *prometheus.CounterVec
	reqConsensusCounter   *prometheus.CounterVec
	reqConsensusTimeGauge *prometheus.GaugeVec
	reqSendCounter        *prometheus.CounterVec
	reqSendTimeGauge      *prometheus.GaugeVec
	reqSubGauge           *prometheus.GaugeVec
	chainOnlineGauge      *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		disabled: true,
	}
}

func (m *Metrics) Init(peerId peer.ID) error {

	m.reqReceivedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_received_count",
		Help:        "Received requests count partitioned by peer_id, chain_id, req_type['bridge_oracle_request'], dst_chain_id",
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "req_type"})

	m.reqConsensusCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_consensus_count",
		Help:        "Consensus for request count partitioned by peer_id, chain_id, req_type[bridge_oracle_request,uptime], status[success,failed]",
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "req_type", "status"})

	m.reqSendCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_send_count",
		Help:        "Received requests count partitioned by peer_id, chain_id, to_chain_id, req_type[bridge_oracle_request,uptime], status[success,failed]",
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "to_chain_id", "req_type", "status"})

	m.reqSubGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_sub",
		Help:        `Subscriptions for request partitioned by peer_id, chain_id, req_type[bridge_oracle_request], action:[subscribe,resubscribe], status[success,failed]`,
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "req_type", "action", "status"})

	m.reqConsensusTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_consensus_time",
		Help:        `Subscriptions for request partitioned by peer_id, chain_id, req_type[bridge_oracle_request,uptime], duration[<5ms,10ms,100ms,500ms,1s,5s,10s,30s,>30s]`,
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "req_type", "duration"})

	m.reqSendTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "req_send_time",
		Help:        `Subscriptions for request partitioned by peer_id, chain_id, to_chain_id, req_type[bridge_oracle_request,uptime], duration[<5ms,10ms,100ms,500ms,1s,5s,10s,30s,>30s]`,
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id", "to_chain_id", "req_type", "duration"})

	m.chainOnlineGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "chain_online",
		Help:        `chain online gauge partitioned by peer_id, chain_id`,
		ConstLabels: prometheus.Labels{"peer_id": peerId.Pretty()},
	}, []string{"chain_id"})

	if err := prometheus.Register(m.reqReceivedCounter); err != nil {

		return fmt.Errorf("register req_received_count error: %w", err)
	} else if err := prometheus.Register(m.reqConsensusCounter); err != nil {

		return fmt.Errorf("register req_consensus_count error: %w", err)
	} else if err := prometheus.Register(m.reqSendCounter); err != nil {

		return fmt.Errorf("register req_send_count error: %w", err)
	} else if err := prometheus.Register(m.reqSubGauge); err != nil {

		return fmt.Errorf("register req_sub error: %w", err)
	} else if err := prometheus.Register(m.reqConsensusTimeGauge); err != nil {

		return fmt.Errorf("register req_consensus_time error: %w", err)
	} else if err := prometheus.Register(m.reqSendTimeGauge); err != nil {

		return fmt.Errorf("register req_send_time error: %w", err)
	} else if err := prometheus.Register(m.chainOnlineGauge); err != nil {

		return fmt.Errorf("register chain_online error: %w", err)
	} else {
		m.disabled = false
		return nil
	}
}

func (m *Metrics) Disable() {
	m.disabled = true
}

func (m *Metrics) ChainMetrics(chainId string) *ChainMetrics {
	return &ChainMetrics{
		m:       m,
		chainId: chainId,
	}
}

func (m *Metrics) RequestMetrics(chainId string, reqType ReqTypeEnum) *RequestMetrics {
	return &RequestMetrics{
		ChainMetrics: *m.ChainMetrics(chainId),
		reqType:      reqType,
	}
}
