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
	"golang.org/x/crypto/sha3"
)

func MakeKeyDir(keysPath string) {
	if !FileExists(keysPath) {
		os.MkdirAll(keysPath, os.ModePerm)
	} else {
		logrus.Warnf("Key directory %s already exists.", keysPath)
	}
}

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

func GenAndSaveECDSAKey(keysPath, prefix string) (ecdsa crypto.PrivKey, err error) {
	nodeKeyFile := keysPath + "/" + prefix + "-ecdsa.key"
	if !FileExists(nodeKeyFile) {
		ecdsa, _, err = crypto.GenerateECDSAKeyPair(rand.Reader)
		if err != nil {
			return nil, err
		}
		err = WriteKey(ecdsa, keysPath, prefix+"-ecdsa")
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Warnf("Key %s exists, reusing it!", nodeKeyFile)
		ecdsa, err = ReadHostKey(nodeKeyFile)
		if err != nil {
			return nil, err
		}
	}

	return
}

func GetOrGenAndSaveECDSAKey(keysPath, prefix string) (hostKey crypto.PrivKey, err error) {
	nodeKeyFile := keysPath + "/" + prefix + "-ecdsa.key"
	if !FileExists(nodeKeyFile) {
		if ecdsa, _, err := crypto.GenerateECDSAKeyPair(rand.Reader); err != nil {

			return nil, err
		} else if err = WriteKey(ecdsa, keysPath, prefix+"-ecdsa"); err != nil {

			return nil, err
		}
	}
	return ReadHostKey(nodeKeyFile)
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

func GetOrGenAndSaveSecp256k1Key(keysPath, prefix string) (pk *ecdsa.PrivateKey, err error) {
	nodeKeyFile := keysPath + "/" + prefix + "-secp256k1.key"
	if !FileExists(nodeKeyFile) {
		if pk, err = ecrypto.GenerateKey(); err != nil {

			return nil, err
		} else if err = ecrypto.SaveECDSA(nodeKeyFile, pk); err != nil {

			return nil, err
		}
	}
	return ecrypto.LoadECDSA(nodeKeyFile)
}

func LoadSecp256k1Key(keysPath, prefix string) (*ecdsa.PrivateKey, error) {
	return ecrypto.LoadECDSA(keysPath + "/" + prefix + "-secp256k1" + ".key")
}

func AddressFromSecp256k1PrivKey(pk *ecdsa.PrivateKey) common.Address {
	return ecrypto.PubkeyToAddress(pk.PublicKey)
}
