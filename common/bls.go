package common

import (
	"crypto/rand"
	"crypto/sha256"
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

func ZeroSignature() BlsSignature {
	return BlsSignature{p: new(bn256.G1).Set(&zeroG1)}
}

func ZeroPublicKey() BlsPublicKey {
	return BlsPublicKey{p: new(bn256.G2).Set(&zeroG2)}
}

func (signature *BlsSignature) IsSet() bool {
	return signature.p != nil
}

func (secretKey BlsPrivateKey) Sign(message []byte) BlsSignature {
	hashPoint := altbn128.G1HashToPoint(message)
	return BlsSignature{p: new(bn256.G1).ScalarMult(hashPoint, secretKey.p)}
}

func (secretKey BlsPrivateKey) Multisign(message []byte, aggPublicKey BlsPublicKey, membershipKey BlsSignature) BlsSignature {
	s := new(bn256.G1).ScalarMult(hashToPointMsg(aggPublicKey.p, message), secretKey.p)
	s.Add(s, membershipKey.p)
	return BlsSignature{p: s}
}

func (signature BlsSignature) Verify(publicKey BlsPublicKey, message []byte) bool {
	hashPoint := altbn128.G1HashToPoint(message)

	a := []*bn256.G1{new(bn256.G1).Neg(signature.p), hashPoint}
	b := []*bn256.G2{&g2, publicKey.p}
	return bn256.PairingCheck(a, b)
}

// VerifyMembershipKeyPart verifies membership key part i ((a⋅pk)×H(P, i))
// against aggregated public key (P) and public key of the party (pk×G)
func (signature BlsSignature) VerifyMembershipKeyPart(aggPublicKey BlsPublicKey, partPublicKey BlsPublicKey, index byte) bool {
	hashPoint := hashToPointIndex(aggPublicKey.p, index)

	a := []*bn256.G1{new(bn256.G1).Neg(signature.p), hashPoint}
	b := []*bn256.G2{&g2, partPublicKey.p}
	return bn256.PairingCheck(a, b)
}

func (signature BlsSignature) VerifyMultisig(allPublicKey BlsPublicKey, publicKey BlsPublicKey, message []byte, bitmask *big.Int) bool {
	sum := new(bn256.G1).Set(&zeroG1)
	mask := new(big.Int).Set(bitmask)
	for index := 0; mask.Sign() != 0; index++ {
		if bitmask.Bit(index) != 0 {
			mask.SetBit(mask, index, 0)
			sum.Add(sum, hashToPointIndex(allPublicKey.p, byte(index)))
		}
	}

	a := []*bn256.G1{new(bn256.G1).Neg(signature.p), hashToPointMsg(allPublicKey.p, message), sum}
	b := []*bn256.G2{&g2, publicKey.p, allPublicKey.p}
	return bn256.PairingCheck(a, b)
}

func (signature *BlsSignature) Aggregate(onemore BlsSignature) {
	signature.p.Add(signature.p, onemore.p)
}

func (pub *BlsPublicKey) Aggregate(onemore BlsPublicKey) {
	pub.p.Add(pub.p, onemore.p)
}

// AggregateBlsSignatures sums the given array of signatures
func AggregateBlsSignatures(sigs []BlsSignature) BlsSignature {
	p := *new(bn256.G1).Set(&zeroG1)
	for _, sig := range sigs {
		p.Add(&p, sig.p)
	}
	return BlsSignature{p: &p}
}

// AggregateBlsPublicKeys calculates P1*A1 + P2*A2 + ...
func AggregateBlsPublicKeys(pubs []BlsPublicKey, anticoefs []big.Int) BlsPublicKey {
	res := *new(bn256.G2).Set(&zeroG2)
	for i := 0; i < len(pubs); i++ {
		res.Add(&res, new(bn256.G2).ScalarMult(pubs[i].p, &anticoefs[i]))
	}
	return BlsPublicKey{p: &res}
}

func (pub BlsPublicKey) Marshal() []byte {
	if pub.p == nil {
		return nil
	}
	return pub.p.Marshal()
}

func (signature BlsSignature) Marshal() []byte {
	if signature.p == nil {
		return nil
	}
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

func GenRandomBlsKey() (BlsPrivateKey, BlsPublicKey) {
	priv, pub, _ := bn256.RandomG2(rand.Reader)
	return BlsPrivateKey{p: priv}, BlsPublicKey{p: pub}
}

// CalculateAntiRogueCoefficients returns an array of bigints used for
// subsequent key aggregations:
//
// Ai = hash(Pi, {P1, P2, ...})
func CalculateAntiRogueCoefficients(pubs []BlsPublicKey) []big.Int {
	as := make([]big.Int, len(pubs))
	data := pubs[0].p.Marshal()
	for i := 0; i < len(pubs); i++ {
		data = append(data, pubs[i].p.Marshal()...)
	}

	for i := 0; i < len(pubs); i++ {
		cur := pubs[i].p.Marshal()
		copy(data[0:len(cur)], cur)
		hash := sha256.Sum256(data)
		as[i].SetBytes(hash[:])
	}
	return as
}

// HashToPointMsg performs "message augmentation": hashes the message and the
// point to the point of G1 curve (a signature)
func hashToPointMsg(p *bn256.G2, message []byte) *bn256.G1 {
	var data []byte
	data = append(data, p.Marshal()...)
	data = append(data, message...)
	return altbn128.G1HashToPoint(data)
}

// HashToPointIndex hashes the aggregated public key (G2 point) and the given
// index (of the signer within a group of signers) to the point in G1 curve (a
// signature)
func hashToPointIndex(pub *bn256.G2, index byte) *bn256.G1 {
	data := make([]byte, 32)
	data[31] = index
	return hashToPointMsg(pub, data)
}

// GenBlsMembershipKeyPart generates the participant signature to be aggregated into membership key
func GenBlsMembershipKeyPart(priv BlsPrivateKey, index byte, aggPub BlsPublicKey, anticoef big.Int) BlsSignature {
	res := new(bn256.G1).ScalarMult(hashToPointIndex(aggPub.p, index), priv.p)
	res.ScalarMult(res, &anticoef)
	return BlsSignature{p: res}
}
