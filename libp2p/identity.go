package libp2p

import (
	"github.com/digiu-ai/p2p-bridge/common"
	"github.com/libp2p/go-libp2p"
	"github.com/sirupsen/logrus"
)

func IdentityFromKey(keyFile string) (identity libp2p.Option, err error) {
	privKey, err := common.ReadHostKey(keyFile)
	if err != nil {
		logrus.Errorf("ERROR GETTING CERT %v", err)
		return
	}
	identity = libp2p.Identity(privKey)
	return
}
