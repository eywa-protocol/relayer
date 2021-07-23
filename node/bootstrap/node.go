package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/runa"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"

	"github.com/libp2p/go-libp2p-core/crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/libp2p"
)

func NodeInit(keysPath, name, listen string, port uint) (err error) {

	keyFile := keysPath + "/" + name + "-rsa.key"
	if common2.FileExists(keyFile) {
		return errors.New("node already initialized! ")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

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
	sentry.AddTags(map[string]string{
		field.PeerId: h.ID().Pretty(),
	})

	_, err = dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	runa.Host(h, cancel, &sync.WaitGroup{})

	return
}
