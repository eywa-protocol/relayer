package bridge

type ChainMetrics struct {
	m       *Metrics
	chainId string
}

func (c *ChainMetrics) ChainOnline() {
	if c.m.disabled {
		return
	}
	c.m.chainOnlineGauge.WithLabelValues(c.chainId).Set(1)
}

func (c *ChainMetrics) ChainOffline() {
	if c.m.disabled {
		return
	}
	c.m.chainOnlineGauge.WithLabelValues(c.chainId).Set(0)
}
