package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile *os.File

	latestBlock     Block
	latestBlockHash Hash
	hasGenesisBlock bool
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

	dbFilePath := getBlockDBFilePath(dataDir)

	f, err := os.OpenFile(dbFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	state := &State{
		Balances:  balances,
		txMempool: make([]Tx, 0),

		dbFile:          f,
		latestBlock:     Block{},
		latestBlockHash: Hash{},
		hasGenesisBlock: false,
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error while scanning block.db file: %w", err)
		}

		blockJSON := scanner.Bytes()

		if len(blockJSON) == 0 {
			break
		}

		var blockFS BlockFS
		err = json.Unmarshal(blockJSON, &blockFS)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshalling JSON to blockFS struct: %w", err)
		}

		err = applyTXs(blockFS.Value.TXs, state)
		if err != nil {
			return nil, fmt.Errorf("error while calculating balances: %w", err)
		}

		state.latestBlock = blockFS.Value
		state.latestBlockHash = blockFS.Key
		state.hasGenesisBlock = true
	}

	return state, nil
}

func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := s.AddBlock(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()
	err := applyBlock(b, pendingState)
	if err != nil {
		return Hash{}, err
	}

	// block is verfied and ready to be added to blockchain
	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFS := BlockFS{
		Key:   blockHash,
		Value: b,
	}

	blockFSJSON, err := json.Marshal(blockFS)
	if err != nil {
		return Hash{}, err
	}

	fmt.Println("Persisting new Block to disk:")
	fmt.Printf("%s\n", blockFSJSON)

	_, err = s.dbFile.Write(append(blockFSJSON, '\n'))
	if err != nil {
		return Hash{}, err
	}

	s.Balances = pendingState.Balances
	s.latestBlockHash = blockHash
	s.latestBlock = b
	s.hasGenesisBlock = true

	return blockHash, nil
}

// applyBlock verfies if block can be added to blockchain
// Block meta data are verfied as well as transactions within (sufficient balances) etc
func applyBlock(b Block, s State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	return applyTXs(b.TXs, &s)
}

func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		if err := applyTx(tx, s); err != nil {
			return err
		}
	}
	return nil
}

func applyTx(tx Tx, state *State) error {
	if tx.IsReward() {
		state.Balances[tx.To] = state.Balances[tx.To] + tx.Value
		return nil
	}

	if tx.Value > state.Balances[tx.From] {
		return fmt.Errorf("invalid transaction. Sender '%s' balance is %d TBB. Tx cost is %d TBB", tx.From, state.Balances[tx.From], tx.Value)
	}

	state.Balances[tx.From] = state.Balances[tx.From] - tx.Value
	state.Balances[tx.To] = state.Balances[tx.To] + tx.Value

	return nil
}

func (s *State) copy() State {
	c := State{}
	c.hasGenesisBlock = s.hasGenesisBlock
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.txMempool = make([]Tx, len(s.txMempool))
	c.Balances = make(map[Account]uint)

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for _, tx := range s.txMempool {
		c.txMempool = append(c.txMempool, tx)
	}

	return c
}

func (s *State) Persist() (Hash, error) {
	latestBlockHash, err := s.latestBlock.Hash()
	if err != nil {
		return Hash{}, err
	}

	block := NewBlock(latestBlockHash, s.latestBlock.Header.Number+1, uint64(time.Now().Unix()), s.txMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFS{blockHash, block}
	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)
	if _, err = s.dbFile.Write(append(blockFsJson, '\n')); err != nil {
		return Hash{}, err
	}

	s.latestBlockHash = blockHash
	s.latestBlock = block
	s.txMempool = []Tx{}

	return latestBlockHash, nil
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LatestBlock().Header.Number + 1
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() error {
	return s.dbFile.Close()
}
