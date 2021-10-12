package bls

import (
	"crypto/sha256"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// CalculateAntiRogueCoefficients returns an array of bigints used for
// subsequent key aggregations:
//
// Ai = hash(Pi, {P1, P2, ...})
func CalculateAntiRogueCoefficients(pubs []PublicKey) []big.Int {
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

// GenerateMembershipKeyPart generates the participant signature to be aggregated into membership key
func GenerateMembershipKeyPart(priv PrivateKey, index byte, aggPub PublicKey, anticoef big.Int) Signature {
	res := new(bn256.G1).ScalarMult(hashToPointIndex(aggPub.p, index), priv.p)
	res.ScalarMult(res, &anticoef)
	return Signature{p: res}
}

// AggregateSignatures sums the given array of signatures
func AggregateSignatures(sigs []Signature) Signature {
	p := *new(bn256.G1).Set(&zeroG1)
	for _, sig := range sigs {
		p.Add(&p, sig.p)
	}
	return Signature{p: &p}
}

// AggregatePublicKeys calculates P1*A1 + P2*A2 + ...
func AggregatePublicKeys(pubs []PublicKey, anticoefs []big.Int) PublicKey {
	res := *new(bn256.G2).Set(&zeroG2)
	for i := 0; i < len(pubs); i++ {
		res.Add(&res, new(bn256.G2).ScalarMult(pubs[i].p, &anticoefs[i]))
	}
	return PublicKey{p: &res}
}
