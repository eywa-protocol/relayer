package common

import (
	"os"

	"gitlab.digiu.ai/blockchainlaboratory/bls-crypto/bls"
)

func ReadBlsPrivateKeyFromFile(fileName string) (priv bls.PrivateKey, err error) {
	text, err := os.ReadFile(fileName)
	if err != nil {
		return
	}
	return bls.UnmarshalPrivateKey(text)
}

func GenAndSaveBlsKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	if !FileExists(nodeKeyFile) {
		prvKey, pubKey := bls.GenerateRandomKey()

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
