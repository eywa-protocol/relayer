package libp2p

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"sort"
)

// LeaderNode use the given input buf to calculate a hash , and consistently choose a node as a master coordinate note
func LeaderNode(msgID string, pIDs []string) (string, error) {
	if len(pIDs) == 0 || len(msgID) == 0 {
		return "", errors.New("invalid input for finding the leader")
	}
	keyStore := make(map[string]string)
	hashes := make([]string, len(pIDs))
	for i, el := range pIDs {
		sum := sha256.Sum256([]byte(msgID + el))
		encodedSum := hex.EncodeToString(sum[:])
		keyStore[encodedSum] = el
		hashes[i] = encodedSum
	}
	sort.Strings(hashes)
	return keyStore[hashes[0]], nil
}

func RelayerLeaderNode(msgID string, pIDs []peer.ID) (peer.ID, error) {
	if len(pIDs) == 0 || len(msgID) == 0 {
		return "", errors.New("invalid input for finding the leader")
	}
	keyStore := make(map[string]peer.ID)
	hashes := make([]string, len(pIDs))
	for i, el := range pIDs {
		sum := sha256.Sum256([]byte(msgID + el.Pretty()))
		encodedSum := hex.EncodeToString(sum[:])
		keyStore[encodedSum] = el
		hashes[i] = encodedSum
	}
	sort.Strings(hashes)
	return keyStore[hashes[0]], nil
}
