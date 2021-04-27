package main

import (
	"crypto/rand"
	"flag"
	"github.com/DigiU-Lab/p2p-bridge/common"
	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	prefix := flag.String("prefix", "", "output file name prefix")
	flag.Parse()

	if err := exportKeys(*prefix); err != nil {
		logrus.Fatal(err)
	}
}

func exportKeys(prefix string) error {
	rsa, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	if err != nil {
		return err
	}
	if err := common.WriteKey(rsa, prefix+"-rsa"); err != nil {
		return err
	}

	ecdsa, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return err
	}
	if err := common.WriteKey(ecdsa, prefix+"-ecdsa"); err != nil {
		return err
	}

	ed, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}
	if err := common.WriteKey(ed, prefix+"-ed25519"); err != nil {
		return err
	}

	sec, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return err
	}
	return common.WriteKey(sec, prefix+"-secp256k1")
}
