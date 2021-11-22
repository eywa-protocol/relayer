package tls

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"io/ioutil"
	"net"
	"time"
)

func StartClient(peerIDString string) error {

	priv, _, err := common2.ReadKeys("../keys/cl1-ecdsa.key", "../keys/cl1-rsa.key")
	if err != nil {
		return err
	}

	peerID, err := peer.IDFromString(peerIDString)
	if err != nil {
		return err
	}

	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return err
	}
	fmt.Printf(" Peer ID: %s\n", id.Pretty())
	tp, err := libp2ptls.New(priv)
	if err != nil {
		return err
	}
	remoteAddr := fmt.Sprintln("localhost:5533")
	fmt.Printf("Dialing %s\n", remoteAddr)
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	fmt.Printf("Dialed raw connection to %s\n", conn.RemoteAddr())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sconn, err := tp.SecureOutbound(ctx, conn, peerID)
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated server: %s\n", sconn.RemotePeer().Pretty())
	data, err := ioutil.ReadAll(sconn)
	if err != nil {
		return err
	}
	fmt.Printf("Received message from server: %s\n", string(data))
	return nil
}
