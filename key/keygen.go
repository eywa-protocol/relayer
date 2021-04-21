package main

import (
	"crypto/rand"
	"flag"
	"github.com/sirupsen/logrus"
	"io/ioutil"

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
	if err := writeKey(rsa, prefix+"-rsa"); err != nil {
		return err
	}

	ecdsa, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return err
	}
	if err := writeKey(ecdsa, prefix+"-ecdsa"); err != nil {
		return err
	}

	ed, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}
	if err := writeKey(ed, prefix+"-ed25519"); err != nil {
		return err
	}

	sec, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return err
	}
	return writeKey(sec, prefix+"-secp256k1")
}

func writeKey(priv crypto.PrivKey, name string) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		logrus.Fatal(err)
	}
	filename := name + ".key"
	logrus.Println("Exporting key to", filename)
	return ioutil.WriteFile("keys/"+filename, privBytes, 0644)
}
