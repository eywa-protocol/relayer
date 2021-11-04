package consensus

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"sync"
	"time"
)

// Consensus is the main struct with all states and data related to consensus process.
/*type Consensus struct {
	// FBFTLog stores the pbft messages and blocks during FBFT process
	FBFTLog *FBFTLog
	// phase: different phase of FBFT protocol: pre-prepare, prepare, commit, finish etc
	phase FBFTPhase
	// current indicates what state a node is in
	current State
	// isBackup declarative the node is in backup mode
	isBackup bool
	// 2 types of timeouts: normal and viewchange
	consensusTimeout map[TimeoutType]*utils.Timeout
	// Commits collected from validators.
	aggregatedPrepareSig *bls_core.Sign
	aggregatedCommitSig  *bls_core.Sign
	prepareBitmap        *bls_cosi.Mask
	commitBitmap         *bls_cosi.Mask

	multiSigBitmap *bls_cosi.Mask // Bitmap for parsing multisig bitmap from validators
	multiSigMutex  sync.RWMutex

	// The blockchain this consensus is working on
	Blockchain *core.BlockChain
	// Minimal number of peers in the shard
	// If the number of validators is less than minPeers, the consensus won't start
	MinPeers   int
	pubKeyLock sync.Mutex
	// private/public keys of current node
	priKey multibls.PrivateKeys
	// the publickey of leader
	LeaderPubKey *bls.PublicKeyWrapper
	// blockNum: the next blockNumber that FBFT is going to agree on,
	// should be equal to the blockNumber of next block
	blockNum uint64
	// Blockhash - 32 byte
	blockHash [32]byte
	// Block to run consensus on
	block []byte
	// Shard Id which this node belongs to
	ShardID uint32
	// IgnoreViewIDCheck determines whether to ignore viewID check
	IgnoreViewIDCheck *abool.AtomicBool
	// consensus mutex
	mutex sync.Mutex
	// mutex for verify new block
	verifyBlockMutex sync.Mutex
	// ViewChange struct
	vc *viewChange
	// Signal channel for proposing a new block and start new consensus
	ReadySignal chan ProposalType
	// Channel to send full commit signatures to finish new block proposal
	CommitSigChannel chan []byte
	// The post-consensus job func passed from Node object
	// Called when consensus on a new block is done
	PostConsensusJob func(*types.Block) error
	// The verifier func passed from Node object
	BlockVerifier VerifyBlockFunc
	// verified block to state sync broadcast
	VerifiedNewBlock chan *types.Block
	// will trigger state syncing when blockNum is low
	BlockNumLowChan chan struct{}
	// Channel for DRG protocol to send pRnd (preimage of randomness resulting from combined vrf
	// randomnesses) to consensus. The first 32 bytes are randomness, the rest is for bitmap.
	PRndChannel chan []byte
	// Channel for DRG protocol to send VDF. The first 516 bytes are the VDF/Proof and the last 32
	// bytes are the seed for deriving VDF
	RndChannel  chan [vdfAndSeedSize]byte
	pendingRnds [][vdfAndSeedSize]byte // A list of pending randomness
	// The p2p host used to send/receive p2p messages
	host p2p.Host
	// MessageSender takes are of sending consensus message and the corresponding retry logic.
	msgSender *MessageSender
	// Used to convey to the consensus main loop that block syncing has finished.
	syncReadyChan chan struct{}
	// Used to convey to the consensus main loop that node is out of sync
	syncNotReadyChan chan struct{}
	// If true, this consensus will not propose view change.
	disableViewChange bool
	// Have a dedicated reader thread pull from this chan, like in node
	SlashChan chan slash.Record
	// How long in second the leader needs to wait to propose a new block.
	BlockPeriod time.Duration
	// The time due for next block proposal
	NextBlockDue time.Time
	// Temporary flag to control whether aggregate signature signing is enabled
	AggregateSig bool

	// TODO (leo): an new metrics system to keep track of the consensus/viewchange
	// finality of previous consensus in the unit of milliseconds
	finality int64
	// finalityCounter keep tracks of the finality time
	finalityCounter int64

	dHelper *downloadHelper
}*/

// Consensus is the main struct with all states and data related to consensus process.
type Consensus struct {
	// isBackup declarative the node is in backup mode
	isBackup      bool
	multiSigMutex sync.RWMutex

	// The blockchain this consensus is working on
	Blockchain *core.BlockChain
	// Minimal number of peers in the shard
	// If the number of validators is less than minPeers, the consensus won't start
	MinPeers   int
	pubKeyLock sync.Mutex
	// private/public keys of current node
	priKey bls.PrivateKey
	// the publickey of leader
	LeaderPubKey *bls.PublicKey
	// blockNum: the next blockNumber that FBFT is going to agree on,
	// should be equal to the blockNumber of next block
	blockNum uint64
	// Blockhash - 32 byte
	blockHash [32]byte
	// Block to run consensus on
	block []byte
	// Shard Id which this node belongs to
	ShardID uint32
	// consensus mutex
	mutex sync.Mutex
	// mutex for verify new block
	verifyBlockMutex sync.Mutex
	// Channel to send full commit signatures to finish new block proposal
	CommitSigChannel chan []byte
	// The post-consensus job func passed from Node object
	// Called when consensus on a new block is done
	PostConsensusJob func(*types.Block) error
	// The verifier func passed from Node object
	// will trigger state syncing when blockNum is low
	BlockNumLowChan chan struct{}
	// Channel for DRG protocol to send pRnd (preimage of randomness resulting from combined vrf
	// randomnesses) to consensus. The first 32 bytes are randomness, the rest is for bitmap.
	PRndChannel chan []byte
	// Used to convey to the consensus main loop that block syncing has finished.
	syncReadyChan chan struct{}
	// Used to convey to the consensus main loop that node is out of sync
	syncNotReadyChan chan struct{}
	// If true, this consensus will not propose view change.
	disableViewChange bool
	// How long in second the leader needs to wait to propose a new block.
	BlockPeriod time.Duration
	// The time due for next block proposal
	NextBlockDue time.Time
	// Temporary flag to control whether aggregate signature signing is enabled
	AggregateSig bool
}
