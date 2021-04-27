package common

import (
	"crypto/rand"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/encoding"
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
	logrus.Print("pubkey", pubkey)
	logrus.Print("privkey", privkey)
	return privkey, pubkey, nil

}

func WriteKey(priv crypto.PrivKey, name string) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		logrus.Fatal(err)
	}
	filename := name + ".key"
	logrus.Println("Exporting key to", filename)
	return ioutil.WriteFile("keys/"+filename, privBytes, 0644)
}

func ReadPointFromFile(fileName string) {
	suite := pairing.NewSuiteBn256()
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	p, err := encoding.ReadHexScalar(suite, file)
	if err != nil {
		panic(err)
	}

	pubKey := suite.Point().Mul(p, nil)
	strPub, err := encoding.PointToStringHex(suite, pubKey)
	if err != nil {
		panic(err)
	}
	logrus.Print(strPub)
}

func CreateBN256Key() (strPub string, err error) {
	suite := pairing.NewSuiteBn256()
	prvKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(prvKey, nil)
	str, err := encoding.ScalarToStringHex(suite, prvKey)
	if err != nil {
		return
	}
	err = os.WriteFile("keys/bn256.key", []byte(str), 0644)
	if err != nil {
		return
	}

	strPub, err = encoding.PointToStringHex(suite, pubKey)
	if err != nil {
		return
	}

	err = os.WriteFile("keys/bn256.pub.key", []byte(strPub), 0644)
	if err != nil {
		panic(err)
	}
	return
}

func CreateRSAKey() {
	pr, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	if err != nil {
		panic(err)
	}
	if err := WriteKey(pr, "key-rsa"); err != nil {
		panic(err)
	}
}
