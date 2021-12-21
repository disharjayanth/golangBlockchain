package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile          *os.File
	latestBlockHash Hash
}

func NewStateFromDisk(dataDir string) (*State, error) {
	if err := initDataDirIfNotExists(dataDir); err != nil {
		return nil, err
	}

	gen, err := LoadGenesis(getGenesisJSONFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)

	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	f, err := os.OpenFile(getBlockDBFilePath(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{
		Balances:  balances,
		txMempool: make([]Tx, 0),

		dbFile:          f,
		latestBlockHash: Hash{},
	}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockJSON := scanner.Bytes()
		var blockFS BlockFS
		err = json.Unmarshal(blockJSON, &blockFS)
		if err != nil {
			return nil, err
		}

		err = state.applyBlock(blockFS.Value)
		if err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFS.Key
	}

	return state, nil
}

func (s *State) applyBlock(b Block) error {
	for _, tx := range b.Txs {
		if err := s.apply(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] = s.Balances[tx.To] + tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}

	s.Balances[tx.From] = s.Balances[tx.From] - tx.Value
	s.Balances[tx.To] = s.Balances[tx.To] + tx.Value

	return nil
}

func (s *State) AddBlock(b Block) error {
	for _, tx := range b.Txs {
		if err := s.Addtx(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) Addtx(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return fmt.Errorf("error while adding transaction to state in Add func: %w", err)
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) Persist() (Hash, error) {
	block := NewBlock(s.latestBlockHash, uint64(time.Now().Unix()), s.txMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{
		Key:   blockHash,
		Value: block,
	}

	blockFsJSON, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, fmt.Errorf("error while marshalling BlockFS to json: %w", err)
	}

	fmt.Println("Persisting new Block to disk:")
	fmt.Printf("\t%s\n", blockFsJSON)

	if _, err = s.dbFile.Write(append(blockFsJSON, '\n')); err != nil {
		return Hash{}, fmt.Errorf("error while writing to disk: %w", err)
	}

	s.latestBlockHash = blockHash

	s.txMempool = []Tx{}

	return s.latestBlockHash, nil
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() error {
	return s.dbFile.Close()
}
