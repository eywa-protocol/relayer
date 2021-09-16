package bridge

import (
	"time"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/prom/base"
)

// ReqTypeEnum request type enum
type ReqTypeEnum int

const (
	_ ReqTypeEnum = iota
	ReqTypeBridge
	ReqTypeUptime
)

// request type string representation
const (
	reqTypeBridge = "bridge_oracle_request"
	reqTypeUptime = "uptime"
)

func (rt ReqTypeEnum) String() string {
	switch rt {
	case ReqTypeBridge:
		return reqTypeBridge
	case ReqTypeUptime:
		return reqTypeUptime
	default:
		return base.Undefined
	}
}

const (
	actionSubscribe   = "subscribe"
	actionResubscribe = "resubscribe"
)

type RequestMetrics struct {
	ChainMetrics
	reqType ReqTypeEnum
}

func (r *RequestMetrics) Received() {
	if r.m.disabled {
		return
	}
	r.m.reqReceivedCounter.WithLabelValues(r.chainId, r.reqType.String()).Inc()
}

func (r *RequestMetrics) ConsensusSuccess() {
	if r.m.disabled {
		return
	}
	r.m.reqConsensusCounter.WithLabelValues(r.chainId, r.reqType.String(), base.Success).Inc()
}

func (r *RequestMetrics) ConsensusFailed() {
	if r.m.disabled {
		return
	}
	r.m.reqConsensusCounter.WithLabelValues(r.chainId, r.reqType.String(), base.Failed).Inc()
}

func (r *RequestMetrics) ConsensusTime(startTime time.Time) {
	if r.m.disabled {
		return
	}
	duration := base.Duration(time.Since(startTime))
	r.m.reqConsensusTimeGauge.WithLabelValues(r.chainId, r.reqType.String(), duration.String()).Inc()
}

func (r *RequestMetrics) SentSuccess(toChainId string) {
	if r.m.disabled {
		return
	}
	r.m.reqSendCounter.WithLabelValues(r.chainId, toChainId, r.reqType.String(), base.Success).Inc()
}

func (r *RequestMetrics) SentFailed(toChainId string) {
	if r.m.disabled {
		return
	}
	r.m.reqSendCounter.WithLabelValues(r.chainId, toChainId, r.reqType.String(), base.Failed).Inc()
}

func (r *RequestMetrics) SendTime(toChainId string, startTime time.Time) {
	if r.m.disabled {
		return
	}
	duration := base.Duration(time.Since(startTime))
	r.m.reqSendTimeGauge.WithLabelValues(r.chainId, toChainId, r.reqType.String(), duration.String()).Inc()
}

func (r *RequestMetrics) SubscribeSuccess() {
	if r.m.disabled {
		return
	}
	r.m.reqSubGauge.WithLabelValues(r.chainId, r.reqType.String(), actionSubscribe, base.Success).Inc()
}

func (r *RequestMetrics) SubscribeFailed() {
	if r.m.disabled {
		return
	}
	r.m.reqSubGauge.WithLabelValues(r.chainId, r.reqType.String(), actionSubscribe, base.Failed).Inc()
}

func (r *RequestMetrics) ResubscribeSuccess() {
	if r.m.disabled {
		return
	}
	r.m.reqSubGauge.WithLabelValues(r.chainId, r.reqType.String(), actionResubscribe, base.Success).Inc()
}

func (r *RequestMetrics) ResubscribeFailed() {
	if r.m.disabled {
		return
	}
	r.m.reqSubGauge.WithLabelValues(r.chainId, r.reqType.String(), actionResubscribe, base.Failed).Inc()
}
