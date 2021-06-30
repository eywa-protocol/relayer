package config

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	err := Load("../.data/bridge.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, chain := range App.Chains {
		fmt.Println(chain.EcdsaKey.PublicKey)
		fmt.Println(chain.BridgeAddress.String())
	}
}
