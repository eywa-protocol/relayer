package crypto

import (
	"testing"

	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
)

func TestBN256Pubkey(t *testing.T) {
	pub, err := common2.GenAndSaveBN256Key(".", "test")
	if err != nil {
		return
	}
	t.Log(pub)
	t.Log(len(pub))
	pubBytes := []byte(pub)
	t.Log(pubBytes)
	t.Log(len(pubBytes))

}
