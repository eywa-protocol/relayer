package libp2p

import (
	"github.com/DigiU-Lab/p2p-bridge/keys"
	"github.com/libp2p/go-libp2p"
	"github.com/sirupsen/logrus"
)

func IdentityFromKey(keyFile string) (identity libp2p.Option, err error) {
	privKey, err := keys.ReadHostKey(keyFile)
	if err != nil {
		logrus.Printf("ERROR GETTING CERT %v", err)
		return
	}
	identity = libp2p.Identity(privKey)
	return
}
