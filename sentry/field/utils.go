package field

import (
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

func ListFromBridgeOracleRequest(r *wrappers.BridgeOracleRequest) logrus.Fields {
	return logrus.Fields{
		RequestId:      r.RequestId,
		RequestType:    r.RequestType,
		CainId:         r.Chainid,
		BridgeAddress:  r.Bridge.String(),
		ReceiveSide:    r.ReceiveSide.String(),
		OppositeBridge: r.OppositeBridge.String(),
	}
}
