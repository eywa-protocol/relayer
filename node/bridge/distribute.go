package bridge

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"sync"
)

type BlockMessage struct {
	Source     int // NodeID of message's source
	MsgType    int // Type of message
	BlockBytes []byte
}

func (n *Node) DistributeBlock(block *types.Block) error {
	blockBytes := block.ToArray()
	msg := BlockMessage{
		Source:     n.Id,
		MsgType:    DistributeBlockPhase,
		BlockBytes: blockBytes,
	}
	msgBytes, _ := json.Marshal(msg)
	n.P2PPubSub.Broadcast(msgBytes)
	hash := block.Hash()
	hashStr := hash.ToHexString()
	logrus.Infof("BROADCASTED BLOCK HASH %v", hashStr)
	return nil
}

func (n *Node) ReceiveAndSaveBlock(wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(n.Ctx)
	defer cancel()
	mutex := &sync.Mutex{}
	defer mutex.Unlock()
	for {
		select {
		case <-ctx.Done():
			break
		default:
			rcvdMsg := n.P2PPubSub.Receive(ctx)
			if rcvdMsg == nil {
				logrus.Info("ReceiveBlock receive return nil")
				break
			}
			wg.Add(1)
			go func(msgBytes *[]byte) {
				defer wg.Done()
				var msg BlockMessage
				if err := json.Unmarshal(*msgBytes, &msg); err != nil {
					logrus.Errorf("ReceiveBlock Unable to decode received message, skipped: %v %d", *msgBytes, n.Id)
					return
				}

				switch msg.MsgType {
				case DistributeBlockPhase:
					block, err := types.BlockFromRawBytes(msg.BlockBytes)
					if err != nil {
						logrus.Errorf("BlockFromRawBytes ERROR: %v", err)
						return
					}
					logrus.Infof("Going to ExecAndSaveBlock %s", block.HashString())
					err = n.Ledger.ExecAndSaveBlock(block)
					if err != nil {
						logrus.Errorf("ExecAndSaveBlock ERROR: %v", err)
						return
					}

				}
			}(rcvdMsg)
		}
	}

	return
}
