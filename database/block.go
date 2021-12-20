package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return fmt.Errorf("error while decoding from hexdecimal to slice of byte: %w", err)
}

type BlockHeader struct {
	Parent Hash
	Time   uint64
}

type Block struct {
	Header BlockHeader
	Txs    []Tx
}

type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

func NewBlock(parent Hash, time uint64, tx []Tx) Block {
	return Block{
		Header: BlockHeader{
			Parent: parent,
			Time:   time,
		},
		Txs: tx,
	}
}

func (b Block) Hash() (Hash, error) {
	jsonBlock, err := json.Marshal(b)
	if err != nil {
		return Hash{}, fmt.Errorf("error while mashalling Block struct to json: %w", err)
	}

	return sha256.Sum256(jsonBlock), nil
}
