package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Snapshot [32]byte

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

	dbFile   *os.File
	snapshot Snapshot
}

func NewStateFromDisk() (*State, error) {
	// Current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch current working directory: %w", err)
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

	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f, Snapshot{}}

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

	err = state.doSnapShot()
	if err != nil {
		return nil, fmt.Errorf("error while creating snapshot at NewStateFromDisk func: %w", err)
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

func (s *State) Persist() (Snapshot, error) {
	// Make a copy of mempool because the s.txMempool will be modified
	// in the loop
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJSON, err := json.Marshal(mempool[i])
		if err != nil {
			return Snapshot{}, fmt.Errorf("error while marshalling tx struct to tx json: %w", err)
		}

		fmt.Println("Persisting new tx into disk")
		fmt.Printf("\t%s\n", txJSON)
		if _, err = s.dbFile.Write(append(txJSON, '\n')); err != nil {
			return Snapshot{}, fmt.Errorf("error while writing tx to tx json file: %w", err)
		}

		err = s.doSnapShot()
		if err != nil {
			return Snapshot{}, fmt.Errorf("error while creating snapshot(hashing contenys of file): %w", err)
		}
		fmt.Printf("New DB Snapshot: %x\n", s.snapshot)

		// Remove the Tx written to a file from the mempool
		s.txMempool = append(s.txMempool[:i], s.txMempool[i+1:]...)
	}

	return s.snapshot, nil
}

func (s *State) Close() error {
	return s.dbFile.Close()
}

func (s *State) doSnapShot() error {
	// Re-read the whole file from the first byte
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error while file seek: %w", err)
	}

	txData, err := ioutil.ReadAll(s.dbFile)
	if err != nil {
		return fmt.Errorf("error while reading tx file: %w", err)
	}
	s.snapshot = sha256.Sum256(txData)

	return nil
}

func (s *State) LatestSnapShot() Snapshot {
	return s.snapshot
}
