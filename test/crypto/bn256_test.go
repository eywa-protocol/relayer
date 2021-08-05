package crypto

import (
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"testing"
)

func TestBN256Pubkey(t *testing.T) {
	_, pub, err := common2.GenAndSaveBN256Key(".", "test")
	if err != nil {
		return
	}
	t.Log(pub)
	t.Log(len(pub))
	pubBytes := []byte(pub)
	t.Log(pubBytes)
	t.Log(len(pubBytes))

}
