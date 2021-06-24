package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"github.com/digiu-ai/p2p-bridge/runa"
	"io/ioutil"
	"math/rand"

	common2 "github.com/digiu-ai/p2p-bridge/common"
	"github.com/digiu-ai/p2p-bridge/libp2p"
	_ "github.com/digiu-ai/p2p-bridge/runa"
	"github.com/libp2p/go-libp2p-core/crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

func NodeInit(keysPath, name, listen string, port uint) (err error) {

	keyFile := keysPath + "/" + name + "-rsa.key"
	if common2.FileExists(keyFile) {
		return errors.New("node already initialized! ")
	}

	r := rand.New(rand.NewSource(int64(4001)))

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	pkData, err := crypto.MarshalPrivateKey(prvKey)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(keyFile, pkData, 0644)
	if err != nil {
		panic(err)
	}

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listen, port))
	if err != nil {
		panic(err)
	}

	h, err := libp2p.NewHostFromRsaKey(prvKey, multiAddr)
	if err != nil {
		panic(err)
	}

	nodeURL := libp2p.WriteHostAddrToConfig(h, keysPath+"/"+name+"-peer.env")

	logrus.Infof("init bootstrap node: %s", nodeURL)

	return
}

func NewNode(keysPath, name, listen string, port uint) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	keyFile := keysPath + "/" + name + "-rsa.key"

	pkData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		logrus.Fatalf("can not read private key file [%s] on error: %v", keyFile, err)
	}

	pk, err := crypto.UnmarshalPrivateKey(pkData)
	if err != nil {
		logrus.Fatalf("unmarshal private key error: %v", err)
	}

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", listen, port))
	if err != nil {
		logrus.Fatalf("create multiaddr error: %v", err)
	}

	h, err := libp2p.NewHostFromRsaKey(pk, multiAddr)
	if err != nil {
		logrus.Fatal(fmt.Errorf("new bootstrap host error: %w", err))
	}

	_, err = dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	runa.Host(h, cancel)

	return
}
