package config

import (
	"fmt"
	"testing"
)

func TestBridgeLoad(t *testing.T) {
	err := LoadBridgeConfig("../.data/bridge.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, chain := range Bridge.Chains {
		fmt.Println(chain.EcdsaKey.PublicKey)
		fmt.Println(chain.BridgeAddress.String())
	}
}
