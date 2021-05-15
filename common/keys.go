package common

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/encoding"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"os"
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

func WriteKey(priv crypto.PrivKey, name string) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		logrus.Fatal(err)
	}
	filename := name + ".key"
	logrus.Infof("Exporting key to %s", filename)
	return ioutil.WriteFile("keys/"+filename, privBytes, 0644)
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CreateBN256Key(name string) (blsAddr common.Address, strPub string, err error) {
	if !fileExists("keys/") {
		os.MkdirAll("keys", os.ModePerm)
	}
	suite := pairing.NewSuiteBn256()

	nodeKeyFile := "keys/" + name + "-bn256.key"

	if !fileExists(nodeKeyFile) {
		logrus.Tracef("CREATING KEYS")

		prvKey := suite.Scalar().Pick(suite.RandomStream())
		pubKey := suite.Point().Mul(prvKey, nil)
		//blsAddr = common.BytesToAddress([]byte(pubKey.String()))
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

		err = os.WriteFile("keys/"+name+"-bn256.pub.key", []byte(strPub), 0644)
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
	blsAddr = BLSAddrFromKeyFile(nodeKeyFile)
	return
}

func BLSAddrFromKeyFile(nodeKeyFile string) (blsAddr common.Address) {
	suite := pairing.NewSuiteBn256()
	p, err := ReadScalarFromFile(nodeKeyFile)
	if err != nil {
		panic(err)
	}
	pubKey := suite.Point().Mul(p, nil)
	strPub, err := encoding.PointToStringHex(suite, pubKey)
	if err != nil {
		panic(err)
	}
	return common.BytesToAddress(Keccak256([]byte(strPub)))
}

func CreateRSAKey(name string) (err error) {
	pr, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	if err != nil {
		return
	}
	if err = WriteKey(pr, name+"key-rsa"); err != nil {
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

func GenECDSAKey(prefix string) (err error) {
	nodeKeyFile := "keys/" + prefix + "-ecdsa.key"
	if !fileExists(nodeKeyFile) {
		ecdsa, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
		if err != nil {
			return err
		}
		err = WriteKey(ecdsa, prefix+"-ecdsa")
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
