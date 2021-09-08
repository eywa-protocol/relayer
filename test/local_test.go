package test

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_Local_SendRequestV2_FromRinkebyToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1111), big.NewInt(1112), testData)
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromMumbaiToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1112), big.NewInt(1111), testData)
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromRinkebyToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1111), big.NewInt(1113), testData)
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromBscToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1113), big.NewInt(1111), testData)
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromBscToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1112), big.NewInt(1113), testData)
	require.NoError(t, err)
}

func Test_Local_SendRequestV2_FromMumbaiToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(1113), big.NewInt(1112), testData)
	require.NoError(t, err)
}
