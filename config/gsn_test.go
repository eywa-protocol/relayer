package config

import (
	"fmt"
	"testing"
)

func TestGsnLoad(t *testing.T) {
	err := LoadGsnConfig("../.data/gsn.yaml")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("TickerInterval:", Gsn.TickerInterval.String())
	for _, chain := range Gsn.Chains {
		fmt.Println(chain.EcdsaKey.PublicKey)
		fmt.Println(chain.ForwarderAddress.String())
	}
}
