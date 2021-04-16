package tls

import (
	"context"
	"flag"
	"fmt"
	"github.com/digiu-ai/p2p-bridge/keys"
	"net"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
)

func StartServer() (id peer.ID, err error) {
	port := flag.Int("p", 5533, "port")
	//keyType := flag.String("key", "ecdsa", "rsa, ecdsa, ed25519 or secp256k1")
	//flag.Parse()
	//
	//priv, err := generateKey(*keyType)
	//if err != nil {
	//	return
	//}
	priv, _, err := keys.ReadKeys("../keys/tls1-ecdsa.key", "../keys/tls1-rsa.key")
	if err != nil {
		return
	}
	id, err = peer.IDFromPrivateKey(priv)
	if err != nil {
		return
	}
	fmt.Printf(" Peer ID: %s\n", id.Pretty())
	tp, err := libp2ptls.New(priv)
	if err != nil {
		return
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		return
	}
	fmt.Printf("Listening for new connections on %s\n", ln.Addr())
	fmt.Printf("Now run the following command in a separate terminal:\n")
	fmt.Printf("\tgo run cmd/tlsdiag.go client -p %d -id %s\n", *port, id.Pretty())

	for {
		conn, err := ln.Accept()
		if err != nil {
			return id, err
		}
		fmt.Printf("Accepted raw connection from %s\n", conn.RemoteAddr())
		go func() {
			if err := handleConn(tp, conn); err != nil {
				fmt.Printf("Error handling connection from %s: %s\n", conn.RemoteAddr(), err)
			}
		}()
	}
	return
}

func handleConn(tp *libp2ptls.Transport, conn net.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sconn, err := tp.SecureInbound(ctx, conn)
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated client: %s\n", sconn.RemotePeer().Pretty())
	fmt.Fprintf(sconn, "Hello client!")
	fmt.Printf("Closing connection to %s\n", conn.RemoteAddr())
	return sconn.Close()
}
