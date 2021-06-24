package common

import (
	"crypto/ecdsa"
	"crypto/rand"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/encoding"
	"golang.org/x/crypto/sha3"
)

// We pass in both the private keys of host and peer.
// We never use the private key of the peer though.
// That's why this function returns the peer's public key.
func ReadKeys(hostKeyFile, peerKeyFile string) (crypto.PrivKey, crypto.PubKey, error) {
	// read the host key
	hostKeyBytes, err := ioutil.ReadFile(hostKeyFile)
	if err != nil {
		return nil, nil, err
	}
	hostKey, err := crypto.UnmarshalPrivateKey(hostKeyBytes)
	if err != nil {
		return nil, nil, err
	}
	// read the peers key
	peerKeyBytes, err := ioutil.ReadFile(peerKeyFile)
	if err != nil {
		return nil, nil, err
	}
	peerKey, err := crypto.UnmarshalPrivateKey(peerKeyBytes)
	if err != nil {
		return nil, nil, err
	}
	return hostKey, peerKey.GetPublic(), nil
}

func ReadHostKey(hostKeyFile string) (hostKey crypto.PrivKey, err error) {
	// read the host key
	hostKeyBytes, err := ioutil.ReadFile(hostKeyFile)
	if err != nil {
		return
	}
	hostKey, err = crypto.UnmarshalPrivateKey(hostKeyBytes)
	if err != nil {
		return
	}
	return
}

func GenPrivPubkey() ([]byte, []byte, error) {
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 2048, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	privkey, err := priv.Raw()
	if err != nil {
		return nil, nil, err

	}
	pubkey, err := pub.Raw()
	if err != nil {
		return nil, nil, err

	}
	return privkey, pubkey, nil

}

func WriteKey(priv crypto.PrivKey, keysPath, name string) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		logrus.Fatal(err)
	}
	filename := name + ".key"
	logrus.Infof("Exporting key to %s", filename)
	return ioutil.WriteFile(keysPath+"/"+filename, privBytes, 0644)
}

func ReadScalarFromFile(fileName string) (p kyber.Scalar, err error) {
	suite := pairing.NewSuiteBn256()
	file, err := os.Open(fileName)
	if err != nil {
		return
	}
	p, err = encoding.ReadHexScalar(suite, file)
	if err != nil {
		return
	}
	return
}

func GenAndSaveBN256Key(keysPath, name string) (blsAddr common.Address, strPub string, err error) {
	if !FileExists(keysPath) {
		os.MkdirAll(keysPath, os.ModePerm)
	}
	suite := pairing.NewSuiteBn256()

	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	if !FileExists(nodeKeyFile) {
		logrus.Tracef("CREATING KEYS")

		prvKey := suite.Scalar().Pick(suite.RandomStream())
		pubKey := suite.Point().Mul(prvKey, nil)
		// blsAddr = common.BytesToAddress([]byte(pubKey.String()))
		str, err := encoding.ScalarToStringHex(suite, prvKey)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(nodeKeyFile, []byte(str), 0644)
		if err != nil {
			panic(err)
		}

		strPub, err = encoding.PointToStringHex(suite, pubKey)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(keysPath+"/"+name+"-bn256.pub.key", []byte(strPub), 0644)
		if err != nil {
			panic(err)
		}
	} else {
		p, err := ReadScalarFromFile(nodeKeyFile)
		if err != nil {
			panic(err)
		}
		pubKey := suite.Point().Mul(p, nil)
		strPub, err = encoding.PointToStringHex(suite, pubKey)
		if err != nil {
			panic(err)
		}

	}
	blsAddr, err = BLSAddrFromKeyFile(nodeKeyFile)
	if err != nil {
		panic(err)
	}

	return
}

func BLSAddrFromKeyFile(nodeKeyFile string) (blsAddr common.Address, err error) {
	suite := pairing.NewSuiteBn256()
	p, err := ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		return
	}
	pubKey := suite.Point().Mul(p, nil)
	strPub, err := encoding.PointToStringHex(suite, pubKey)
	if err != nil {
		return
	}
	blsAddr = common.BytesToAddress(Keccak256([]byte(strPub)))
	return
}

func CreateRSAKey(keysPath, name string) (err error) {
	pr, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	if err != nil {
		return
	}
	if err = WriteKey(pr, keysPath, name+"key-rsa"); err != nil {
		return
	}
	return
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func GenAndSaveECDSAKey(keysPath, prefix string) (err error) {
	nodeKeyFile := keysPath + "/" + prefix + "-ecdsa.key"
	if !FileExists(nodeKeyFile) {
		ecdsa, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
		if err != nil {
			return err
		}
		err = WriteKey(ecdsa, keysPath, prefix+"-ecdsa")
		if err != nil {
			return err
		}
	} else {
		_, err = ReadHostKey(nodeKeyFile)
		if err != nil {
			return err
		}
	}

	return
}

func AddressFromPrivKey(skey string) (address common.Address) {
	privateKey, err := ToECDSAFromHex(skey)
	if err != nil {
		logrus.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logrus.Fatal("error casting public key to ECDSA")
	}
	address = ecrypto.PubkeyToAddress(*publicKeyECDSA)
	return

}
