package test_utils

import (
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/consensus/model"
	"sync"
	"testing"
)

func runNode(node *model.Node, stop int, wg *sync.WaitGroup) {
	defer wg.Done()
	node.WaitForMsg(stop)
}

// StartTest is used for starting tlc nodes
func StartTest(nodes []*model.Node, stop int, fails int) {
	wg := &sync.WaitGroup{}

	for _, node := range nodes {
		wg.Add(1)
		go node.Advance(0)
	}
	for _, node := range nodes {
		wg.Add(1)
		go runNode(node, stop, wg)
	}
	//wg.Add(-fails)
	wg.Wait()
	fmt.Println("The END")
}

func LogOutput(t *testing.T, nodes []*model.Node) {
	for i := range nodes {
		t.Logf("nodes: %d , TimeStep : %d", i, nodes[i].TimeStep)
		model.Logger1.Printf("%d,%d\n", i, nodes[i].TimeStep)
	}
}
