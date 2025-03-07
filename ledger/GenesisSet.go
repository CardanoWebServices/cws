package ledger

import (
	"encoding/base64"
	"log"
	"os"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/sha3"
)

const GENESIS_ENV_VAR_NAME = "CWS_GENESIS"

type GenesisSet struct {
	Actions []Action
}

func NewGenesisSet(actions ...Action) *GenesisSet {
	return &GenesisSet{actions}
}

func DecodeGenesisSet(bytes []byte) (*GenesisSet, error) {
	lst := [][]byte{}
	err := cbor.Unmarshal(bytes, &lst)

	if err != nil {
		return nil, err
	}

	n := len(lst)
	changes := make([]Action, n)

	for i := 0; i < n; i++ {
		c, err := DecodeAction(lst[i])

		if err != nil {
			return nil, err
		}

		changes[i] = c
	}

	g := GenesisSet{changes}

	return &g, nil
}

func LookupGenesisSet() *GenesisSet {
	str, exists := os.LookupEnv(GENESIS_ENV_VAR_NAME)

	if !exists {
		log.Fatal(GENESIS_ENV_VAR_NAME + " is not set")
	}

	decodedBytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(str)
	if err != nil {
		log.Fatal(err)
	}

	g, err := DecodeGenesisSet(decodedBytes)

	if err != nil {
		log.Fatal(err)
	}

	return g
}

// is encoded as list of bytes
func (g *GenesisSet) Encode() []byte {
	n := len(g.Actions)
	lst := make([][]byte, n)

	for i, a := range(g.Actions) {
		h := NewActionHelper(a)
		lst[i] = h.Encode()
	}

	bytes, err := cbor.Marshal(lst)

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func (g *GenesisSet) EncodeBase64() string {
	bytes := g.Encode()
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
}

func (g *GenesisSet) Hash() ChangeSetHash {
	return sha3.Sum256(g.Encode())
}