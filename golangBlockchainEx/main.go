package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/disharjayanth/golangBlockchain/database"
)

type PendingBlock struct {
	parent database.Hash
	number uint64
	time   uint64
	txs    []database.Tx
}

func NewPendingBlock(parent database.Hash, number uint64, time uint64, txs []database.Tx) PendingBlock {
	return PendingBlock{
		parent: parent,
		number: number,
		time:   time,
		txs:    txs,
	}
}

func createRandomPendingBlock() PendingBlock {
	return NewPendingBlock(database.Hash{}, 0, uint64(time.Now().Unix()), []database.Tx{
		database.NewTx("andrej", "andrej", 3, ""),
		database.NewTx("andrej", "andrej", 700, "reward"),
	})
}

func generateNonce() uint64 {
	return rand.Uint64()
}

func Mine(ctx context.Context, pb PendingBlock) (database.Block, error) {
	start := time.Now()
	hash := database.Hash{}
	attempt := 0
	var block database.Block
	var nonce uint64

	for !database.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			err := fmt.Errorf("mining cancelled %s", ctx.Err())
			return database.Block{}, err
		default:
		}

		attempt++
		nonce = generateNonce()

		if attempt%1000000 == 0 || attempt == 1 {
			fmt.Println("Mining pending txs "+strconv.Itoa(len(pb.txs))+". Attempt ", attempt)
		}

		block = database.NewBlock(pb.parent, pb.number, nonce, pb.time, pb.txs)
		blockHash, err := block.Hash()
		if err != nil {
			return database.Block{}, fmt.Errorf("couldn't mine block %s", err.Error())
		}

		hash = blockHash
	}

	fmt.Println("Mined new block using pow", hash)
	fmt.Println("Height of new block:", block.Header.Number)
	fmt.Println("Nonce:", block.Header.Nonce)
	fmt.Println("Block created at:", block.Header.Time)
	fmt.Println("Parent:", block.Header.Parent)

	fmt.Println("Attempt:", attempt)
	fmt.Println("Since:", time.Since(start))

	return block, nil
}

func main() {
	pendingBlock := createRandomPendingBlock()

	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		log.Fatalln(err)
	}

	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("minedBlockHash:", minedBlockHash)
	fmt.Println("minedBlockHash in hex:", minedBlockHash.Hex())
}
