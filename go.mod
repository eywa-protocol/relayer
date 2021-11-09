module gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge

go 1.16

replace gitlab.digiu.ai/blockchainlaboratory/wrappers => ./external/eth-contracts/wrappers/

replace gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge => ./

replace gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain => ./external/eywa-overhead-chain/

require (
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/certifi/gocertifi v0.0.0-20210507211836-431795d63e8d // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.11
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/eywa-protocol/bls-crypto v0.1.2
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/libp2p/go-flow-metrics v0.0.3
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
	github.com/multiformats/go-multiaddr v0.3.2
	github.com/multiformats/go-multihash v0.0.15
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain v0.0.0-20211018150134-95faf4c0fee4
	gitlab.digiu.ai/blockchainlaboratory/wrappers v0.0.0-20210610095807-6b6bd5e43077
	go.dedis.ch/kyber/v3 v3.0.13
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/sys v0.0.0-20211020174200-9d6173849985 // indirect
	google.golang.org/protobuf v1.26.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
