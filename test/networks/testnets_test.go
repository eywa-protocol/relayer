package networks

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

func Test_SendRequestV2_FromRinkebyToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(4), big.NewInt(80001), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(80001), big.NewInt(4), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_SendRequestV2_FromRinkebyToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(4), big.NewInt(97), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(97), big.NewInt(4), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(97), big.NewInt(80001), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(80001), big.NewInt(97), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}
