package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

const (
	gasTypeNet  = "from_net"
	gasTypeUsed = "used"
)

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

func (c *ChainMetrics) gasPrice2Float64(gasPrice *big.Int) float64 {
	floatGasPrice, _ := new(big.Float).SetInt(gasPrice).Float64()
	return floatGasPrice
}

func (c *ChainMetrics) GasPriceFromNet(gasPrice *big.Int) {
	if c.m.disabled {
		return
	}
	c.m.chainGasPriceGauge.WithLabelValues(c.chainId, gasTypeNet).Set(c.gasPrice2Float64(gasPrice))
}

func (c *ChainMetrics) GasPriceUsed(gasPrice *big.Int) {
	if c.m.disabled {
		return
	}
	c.m.chainGasPriceGauge.WithLabelValues(c.chainId, gasTypeUsed).Set(c.gasPrice2Float64(gasPrice))
}

func (c *ChainMetrics) NodeBalance(balanceWei *big.Int) {

	ethBalance, _ := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(params.Ether)).Float64()
	c.m.nodeBalanceGauge.WithLabelValues(c.chainId).Set(ethBalance)
}
