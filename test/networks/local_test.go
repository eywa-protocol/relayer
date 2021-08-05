package networks

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

func Test_Local_SendRequestV2_FromRinkebyToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1111), big.NewInt(1112), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromMumbaiToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1112), big.NewInt(1111), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromRinkebyToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1111), big.NewInt(1113), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromBscToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1113), big.NewInt(1111), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromBscToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1112), big.NewInt(1113), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromMumbaiToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1113), big.NewInt(1112), big.NewInt(rand.Int63()))
	require.NoError(t, err)
}
