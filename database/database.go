package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Account string

func NewAccount(value string) Account {
	return Account(value)
}

type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

func (t Tx) IsReward() bool {
	return t.Data == "reward"
}

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile *os.File
}

func NewStateFromDisk() (*State, error) {
	// Current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch current working directory %w", err)
	}

	genFilePath := filepath.Join(cwd, "database", "genesisdb.json")
	genesis, err := LoadGenesis(genFilePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading genesis block: %w", err)
	}

	balances := make(map[Account]uint)

	for account, balance := range genesis.Balances {
		balances[account] = balance
	}

	txDBFilePath := filepath.Join(cwd, "database", "tx.db")
	f, err := os.OpenFile(txDBFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("error while opening txdb file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f}

	// Iterate over each line in tx.db file
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, fmt.Errorf("error while scanning tx.db transactions: %w", err)
		}

		// convert tx json into tx struct
		var tx Tx
		json.Unmarshal(scanner.Bytes(), &tx)

		// Rebuild the state (user balances)
		// as a series of events
		if err = state.apply(tx); err != nil {
			return nil, fmt.Errorf("error while update state: %w", err)
		}
	}

	return state, nil
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

func (s *State) Add(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return fmt.Errorf("error while adding transaction to state in Add func: %w", err)
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) Persist() error {
	// Make a copy of mempool because the s.txMempool will be modified
	// in the loop
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJSON, err := json.Marshal(mempool[i])
		if err != nil {
			return fmt.Errorf("error while marshalling tx struct to tx json: %w", err)
		}

		if _, err = s.dbFile.Write(append(txJSON, '\n')); err != nil {
			return fmt.Errorf("error while writing tx to tx json file: %w", err)
		}

		// Remove the Tx written to a file from the mempool
		s.txMempool = s.txMempool[1:]
	}

	return nil
}

func (s *State) Close() error {
	return s.dbFile.Close()
}
