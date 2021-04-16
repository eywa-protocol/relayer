package keyset

import (
	"github.com/libp2p/go-libp2p-core/test"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

type Keyset struct {
	sk crypto.PrivKey
	pk crypto.PubKey
	/*publcKeyHashBytes [32]byte
	publcKeyHashBase58 string*/
}

func NewKeySet() (keyset *Keyset, err error) {
	sk, pk, err := test.RandTestKeyPair(crypto.Secp256k1, 2048)
	if err != nil {
		return
	}

	keyset = &Keyset{}
	keyset.pk = pk
	keyset.sk = sk
	return
}

//	sk:                 sk,
//	pk:                 pk/*,
//	publcKeyHashBytes:  [32]byte{},
//	publcKeyHashBase58: "",*/
//}
//var err error

//b, err := ks.pk.Bytes()
//if err != nil {
//	return err
//}
//ks.publcKeyHashBytes = sha3.Sum256(b)
////ks.publcKeyHashBase58 = b58.Encode(ks.publcKeyHashBytes)
//return nil
//}
