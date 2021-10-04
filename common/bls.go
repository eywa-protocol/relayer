package common

import (
	"crypto/rand"
	"math/big"
	"os"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/altbn128"
	"github.com/sirupsen/logrus"
)

var (
	zeroBigInt = *big.NewInt(0)
	oneBigInt  = *big.NewInt(1)
	g2         = *new(bn256.G2).ScalarBaseMult(&oneBigInt)  // Generator point of G2 group
	zeroG1     = *new(bn256.G1).ScalarBaseMult(&zeroBigInt) // Zero point in G2 group
	zeroG2     = *new(bn256.G2).ScalarBaseMult(&zeroBigInt) // Zero point in G2 group
	EmptyMask  = zeroBigInt
)

type BlsPrivateKey struct {
	p *big.Int
}

type BlsPublicKey struct {
	p *bn256.G2
}

type BlsSignature struct {
	p *bn256.G1
}

func ReadBlsPrivateKeyFromFile(fileName string) (priv BlsPrivateKey, err error) {
	text, err := os.ReadFile(fileName)
	if err != nil {
		return
	}

	p := new(big.Int)
	err = p.UnmarshalJSON(text)
	if err != nil {
		return
	}
	return BlsPrivateKey{p: p}, nil
}

func (priv BlsPrivateKey) PublicKey() BlsPublicKey {
	return BlsPublicKey{p: new(bn256.G2).ScalarBaseMult(priv.p)}
}

func ReadBlsPublicKey(raw []byte) (BlsPublicKey, error) {
	p := new(bn256.G2)
	_, err := p.Unmarshal(raw)
	return BlsPublicKey{p: p}, err
}

func GenAndSaveBlsKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	if !FileExists(nodeKeyFile) {
		logrus.Tracef("CREATING KEYS")

		prvKey, pubKey, err := bn256.RandomG2(rand.Reader)
		if err != nil {
			return nil, err
		}

		str, err := prvKey.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(nodeKeyFile, []byte(str), 0644)
		if err != nil {
			return nil, err
		}

		strPub = pubKey.Marshal()
		err = os.WriteFile(keysPath+"/"+name+"-bn256.pub.key", strPub, 0644)
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Warnf("Key %s exists, reusing it!", nodeKeyFile)
		strPub, err = LoadBlsPublicKey(keysPath, name)
	}
	return
}

func LoadBlsPublicKey(keysPath, name string) (strPub []byte, err error) {
	nodeKeyFile := keysPath + "/" + name + "-bn256.key"

	p, err := ReadBlsPrivateKeyFromFile(nodeKeyFile)
	if err != nil {
		return
	}
	strPub = p.PublicKey().p.Marshal()
	return
}

func (signature *BlsSignature) Clear() {
	signature.p = nil
}

func (secretKey BlsPrivateKey) Sign(message []byte) BlsSignature {
	hashPoint := altbn128.G1HashToPoint(message)
	return BlsSignature{p: new(bn256.G1).ScalarMult(hashPoint, secretKey.p)}
}

func (signature BlsSignature) Verify(publicKey BlsPublicKey, message []byte) bool {
	hashPoint := altbn128.G1HashToPoint(message)

	a := []*bn256.G1{new(bn256.G1).Neg(signature.p), hashPoint}
	b := []*bn256.G2{&g2, publicKey.p}

	return bn256.PairingCheck(a, b)
}

func AggregateBlsSignatures(sigs []BlsSignature, mask *big.Int) BlsSignature {
	p := zeroG1
	for i, sig := range sigs {
		if mask.Bit(i) != 0 {
			p.Add(&p, sig.p)
		}
	}
	return BlsSignature{p: &p}
}

func AggregateBlsPublicKeys(pubs []BlsPublicKey, mask *big.Int) BlsPublicKey {
	p := zeroG2
	for i, pub := range pubs {
		if mask.Bit(i) != 0 {
			p.Add(&p, pub.p)
		}
	}
	return BlsPublicKey{p: &p}
}

func (signature BlsSignature) Marshal() []byte {
	return signature.p.Marshal()
}

func UnmarshalBlsSignature(raw []byte) (BlsSignature, error) {
	p := new(bn256.G1)
	_, err := p.Unmarshal(raw)
	return BlsSignature{p: p}, err
}

func UnmarshalBlsPublicKey(raw []byte) (BlsPublicKey, error) {
	p := new(bn256.G2)
	_, err := p.Unmarshal(raw)
	return BlsPublicKey{p: p}, err
}
