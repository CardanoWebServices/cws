package ledger

import (
	"encoding/base64"
	"errors"
	"github.com/btcsuite/btcutil/bech32"
	"golang.org/x/crypto/blake2b"
	"log"
)

func StringifyHumanReadableBytes(prefix string, bs []byte) string {
	conv, err := bech32.ConvertBits(bs, 8, 5, true)
	if err != nil {
		log.Fatal(err)
	}

	str, err := bech32.Encode(prefix, conv)
	if err != nil {
		log.Fatal(err)
	}

	return str
}

func StringifyCompactBytes(bs []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bs)
}

func ParseHumanReadableBytes(str string, expectedPrefix string) ([]byte, error) {
	prefix, bs, err := bech32.Decode(str)

	if err != nil {
		return nil, err
	}

	if prefix != expectedPrefix {
		return nil, errors.New("unexpected bech32 prefix " + prefix)
	}

	return bech32.ConvertBits(bs, 5, 8, false)
}

func ParseCompactBytes(str string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(str)
}

func DigestCompact(bs []byte) []byte {
	hasher, err := blake2b.New(16, nil)
	if err != nil {
		log.Fatal(err)
	}

	hasher.Write(bs)
	hash := hasher.Sum(nil)

	return hash
}
