package networks

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_SendRequestV2ALL(t *testing.T) {
	Test_SendRequestV2_FromRinkebyToMumbai(t)
	Test_SendRequestV2_FromMumbaiToRinkeby(t)
	Test_SendRequestV2_FromRinkebyToBsc(t)
	Test_SendRequestV2_FromBscToRinkeby(t)
	Test_SendRequestV2_FromBscToMumbai(t)
	Test_SendRequestV2_FromMumbaiToBsc(t)
	Test_SendRequestV2_FromAwaxToHeco(t)
	Test_SendRequestV2_FromHecoToAwax(t)
}




func Test_SendRequestV2_FromRinkebyToMumbai(t *testing.T) { 
	SendRequestV2FromChainToChain(t, big.NewInt(4), big.NewInt(80001), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(80001), big.NewInt(4), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromRinkebyToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(4), big.NewInt(97), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToRinkeby(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(97), big.NewInt(4), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromBscToMumbai(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(97), big.NewInt(80001), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromMumbaiToBsc(t *testing.T) {
	SendRequestV2FromChainToChain(t, big.NewInt(80001), big.NewInt(97), testData)
	require.NoError(t, err)
}

func Test_SendRequestV2_FromAwaxToHeco(t *testing.T) {
	        SendRequestV2FromChainToChain(t, big.NewInt(43113), big.NewInt(256), testData)
		        require.NoError(t, err)
		}

func Test_SendRequestV2_FromHecoToAwax(t *testing.T) {
	        SendRequestV2FromChainToChain(t, big.NewInt(256), big.NewInt(43113), testData)
		        require.NoError(t, err)
		}
