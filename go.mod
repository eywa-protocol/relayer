module github.com/DigiU-Lab/p2p-bridge

go 1.16

replace github.com/DigiU-Lab/p2p-bridge => ./
replace github.com/DigiU-Lab/eth-contracts-go-wrappers => ../eth-contracts/wrappers

require (
	github.com/DigiU-Lab/eth-contracts-go-wrappers v0.0.0-20210419095758-51b134180ee0
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/ethereum/go-ethereum v1.10.2
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/mux v1.8.0
	github.com/ipfs/go-datastore v0.4.5
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-circuit v0.4.0
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-gorpc v0.1.2
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-kad-dht v0.11.1
	github.com/libp2p/go-libp2p-net v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.2.6
	github.com/libp2p/go-libp2p-protocol v0.1.0
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-quic-transport v0.10.0
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/libp2p/go-ws-transport v0.4.0
	github.com/linkpoolio/bridges v0.0.0-20200226172010-aa6f132d477e
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	go.dedis.ch/kyber/v3 v3.0.13
	gopkg.in/urfave/cli.v1 v1.20.0
)
