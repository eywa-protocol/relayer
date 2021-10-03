package common

import (
	"crypto/rand"
	"math/big"
	"os"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/sirupsen/logrus"
)

type BlsPrivateKey struct {
	p *big.Int
}

type BlsPublicKey struct {
	p *bn256.G2
}

type BlsSignature struct {
	p *bn256.G1
}

func ReadBlsPrivateKeyFromFile(fileName string) (priv BlsPrivateKey, err error) {
	text, err := os.ReadFile(fileName)
	if err != nil {
		return
	}

	p := new(big.Int)
	err = p.UnmarshalJSON(text)
	if err != nil {
		return
	}
	return BlsPrivateKey{p: p}, nil
}

func (priv BlsPrivateKey) PublicKey() BlsPublicKey {
	return BlsPublicKey{p: new(bn256.G2).ScalarBaseMult(priv.p)}
}

func ReadBlsPublicKey(raw []byte) (BlsPublicKey, error) {
	p := new(bn256.G2)
	_, err := p.Unmarshal(raw)
	return BlsPublicKey{p: p}, err
}

func GenAndSaveBlsKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	if !FileExists(nodeKeyFile) {
		logrus.Tracef("CREATING KEYS")

		prvKey, pubKey, err := bn256.RandomG2(rand.Reader)
		if err != nil {
			return nil, err
		}

		str, err := prvKey.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(nodeKeyFile, []byte(str), 0644)
		if err != nil {
			return nil, err
		}

		strPub = pubKey.Marshal()
		err = os.WriteFile(keysPath+"/"+name+"-bn256.pub.key", strPub, 0644)
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Warnf("Key %s exists, reusing it!", nodeKeyFile)
		strPub, err = LoadBlsPublicKey(keysPath, name)
	}
	return
}

func LoadBlsPublicKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	p, err := ReadBlsPrivateKeyFromFile(nodeKeyFile)
	if err != nil {
		return
	}
	strPub = p.PublicKey().p.Marshal()
	return
}
