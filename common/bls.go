package common

import (
	"crypto/rand"
	"os"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/crypto/bls"
)

func ReadBlsPrivateKeyFromFile(fileName string) (priv bls.PrivateKey, err error) {
	text, err := os.ReadFile(fileName)
	if err != nil {
		return
	}
	return bls.UnmarshalBlsPrivateKey(text)
}

func GenAndSaveBlsKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	if !FileExists(nodeKeyFile) {
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
	strPub = p.PublicKey().Marshal()
	return
}
