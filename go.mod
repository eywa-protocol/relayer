module github.com/digiu-ai/p2p-bridge

go 1.16

replace github.com/digiu-ai/wrappers => ./external/eth-contracts/wrappers/

replace github.com/digiu-ai/p2p-bridge => ./

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/digiu-ai/wrappers v0.0.0-20210610095807-6b6bd5e43077
	github.com/ethereum/go-ethereum v1.10.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/ipfs/go-datastore v0.4.5
	github.com/libp2p/go-libp2p v0.14.2
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-gorpc v0.1.3
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-kad-dht v0.12.1
	github.com/libp2p/go-libp2p-peerstore v0.2.7
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-quic-transport v0.11.0
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/libp2p/go-ws-transport v0.4.0
	github.com/linkpoolio/bridges v0.0.0-20200226172010-aa6f132d477e
	github.com/multiformats/go-multiaddr v0.3.2
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	go.dedis.ch/kyber/v3 v3.0.13
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
