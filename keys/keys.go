package keys

import (
	"crypto/rand"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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
