package sim

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	common2 "gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/common"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
)

func Test_AddNodeSignRawTx(t *testing.T) {

	blsPubKey = string(GenRandomBytes(256))
	randWallet = common.BytesToAddress(GenRandomBytes(20))
	pool = common.BytesToAddress(GenRandomBytes(20))
	bridgeOwner = common.BytesToAddress(GenRandomBytes(20))

	node := wrappers.NodeRegistryNode{
		Owner:         bridgeOwner,
		Pool:          pool,
		NodeIdAddress: randWallet,
		BlsPubKey:     blsPubKey,
		NodeId:        big.NewInt(0),
	}

	testForwardCreateNodeRegistryABIPacked, err = nodeRegistryABI.Pack("addNode", &node)
	if err != nil {
		panic(err)
	}

	countBefore := getNodesCount()
	signedTx2, err := common2.RawSimTx(backend, owner, &nodeRegistryAddress, testForwardCreateNodeRegistryABIPacked)
	require.NoError(t, err)

	v, r, s := signedTx2.RawSignatureValues()
	t.Log(v)
	t.Log(r)
	t.Log(s)

	err = backend.SendTransaction(context.Background(), signedTx2)
	require.NoError(t, err)
	backend.Commit()
	countAfter := getNodesCount()
	t.Log(countBefore)
	t.Log(countAfter)
	require.True(t, (countAfter-countBefore == 1))

}

func Test_AddNodeSimple(t *testing.T) {
	countBefore := getNodesCount()
	_, err := nodeRegistry.AddNode(
		owner,
		createNodeRegistryData)
	require.NoError(t, err)
	backend.Commit()
	countAfter := getNodesCount()
	t.Log(countBefore)
	t.Log(countAfter)
	require.True(t, (countAfter-countBefore == 1))

}

/*func Test_AddNodeThroughForwarder1(t *testing.T) {
	qwe := GenRandomBytes(20)
	createNodeData.blsPubKey = qwe
	createNodeData.nodeIdAddress = common.BytesToAddress(qwe)
	createNodeData.nodeWallet = common.BytesToAddress(qwe)

	_inputData, err := nodeListABI.Pack("addNode",
		createNodeData.nodeWallet,
		createNodeData.nodeIdAddress,
		createNodeData.blsPubKey)
	require.NoError(t, err)

	nonce, err := forwarder.GetNonce(&bind.CallOpts{}, opts.From)
	require.NoError(t, err)

	forwarederRequest := &wrappers.IForwarderForwardRequest{
		From:  opts.From,
		To:    nodeListAddress,
		Value: big.NewInt(0),
		Gas:   big.NewInt(1e6),
		Nonce: nonce,
		Data:  _inputData,
	}

	step256 := new(uint256.Int)
	step256.SetFromBig(nonce)

	addr, _ := common.NewMixedcaseAddressFromString("0x0011223344556677889900112233445566778899")

	gas := big.NewInt(rand.Int63())
	forwarderRequestTypedData := getForwarderSignerData(_inputData, *nonce, *gas)

	//domainSeparator, err := d.HashStruct("EIP712Domain", d.Domain.Map())

	sigTyped, _, err := common2.SignTypedData(*key, *addr, forwarderRequestTypedData)
	require.NoError(t, err)

	domainSeparator, err := forwarderRequestTypedData.HashStruct("EIP712Domain", forwarderRequestTypedData.Domain.Map())
	//signedTx2, err := common2.RawTx(*client.EthClient, opts, &nodeListAddress, _inputData)
	//t.Log(len(domainSeparator))
	//require.NoError(t, err)
	//
	//v, r, s := signedTx2.RawSignatureValues()
	//t.Log(v, r, s )

	//hashData := crypto.Keccak256(_inputData)
	//sig, err := crypto.Sign(hashData[:], client.EcdsaKey)
	//t.Log(common2.BytesToHex(sig))
	//signedTx2.

	//err = client.EthClient.SendTransaction(context.Background(), signedTx2)

	require.NoError(t, err)

	//qwe, _ = signedTx2.MarshalBinary()
	t.Log("forwarederRequest.To", forwarederRequest.To)
	dsep, _ := common2.FromHex(domainSeparator.String())
	reqType, err := forwarder.GENERICPARAMS(&bind.CallOpts{})
	require.NoError(t, err)
	t.Log(reqType)
	forwardRequestType := fmt.Sprintf("ForwardRequest(%s)", reqType)
	reqTypeHash := crypto.Keccak256([]byte(forwardRequestType))
	reqTypeHashBytes32, err := common2.BytesToBytes32(reqTypeHash)
	//reqTypeHashBytes32, err := common2.BytesToBytes32(sigTypedHash)
	require.NoError(t, err)
	//err = forwarder.Verify(&bind.CallOpts{}, forwarederRequest, dsep , reqTypeHashBytes32, nil, sigTyped)
	//require.NoError(t, err)
	//t.Log(sigTypedHash.String())
	//tx, err := forwarder.RegisterRequestType(opts, sigTypedHash.String(), ")")
	//require.NoError(t, err)
	//require.NotNil(t, tx)
	//status, recaipt := helpers.WaitForBlockCompletation(client.EthClient, tx.Hash().String())
	//t.Log(recaipt.Logs)
	//t.Log(status)
	//time.Sleep(2 * time.Second)
	//q := []byte (sigTyped)
	////0x0 << &q
	r, s, v := decodeSignature(sigTyped)
	t.Log(r, s, v)
	t.Log(sigTyped[64])
	//rBytes := s.Bytes()
	//sBytes := s.Bytes()
	//vBytes := s.Bytes()

	sB := new(big.Int).SetBytes(s.Bytes())
	fmt.Printf("OK: y=%s (bytes=%#v)\n", sB, sB.Bytes())

	//
	//qwe := append(rBytes, sBytes)
	asd := r.String() + s.String() + v.String()
	t.Log(len(sigTyped), len(asd), len(sB.Bytes()))

	tx, err := forwarder.Execute(opts, *forwarederRequest, dsep, reqTypeHashBytes32, nil, sigTyped)
	require.NoError(t, err)
	require.NotNil(t, tx)
	//status, recaipt := helpers.WaitForBlockCompletation(backend, tx.Hash().String())
	//t.Log(recaipt.Logs)
	//t.Log(status)
	//time.Sleep(20 * time.Second)
}
*/
