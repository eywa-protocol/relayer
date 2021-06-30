package bridge

import (
	"context"
	"github.com/digiu-ai/p2p-bridge/config"
)

func NewNodeWithClients(path string) (node *Node, err error) {
	loadConfigSetKeysAndCheck(path)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	node = &Node{
		Ctx: ctx,
	}
	c1, c2, c3, err := getEthClients()
	if err != nil {
		return
	}
	node.Client1, err = setNodeEthClient(c1, config.Config.BRIDGE_NETWORK1, config.Config.NODELIST_NETWORK1, config.Config.ECDSA_KEY_1)
	node.Client2, err = setNodeEthClient(c2, config.Config.BRIDGE_NETWORK2, config.Config.NODELIST_NETWORK2, config.Config.ECDSA_KEY_2)
	node.Client3, err = setNodeEthClient(c3, config.Config.BRIDGE_NETWORK3, config.Config.NODELIST_NETWORK3, config.Config.ECDSA_KEY_3)
	return
}
